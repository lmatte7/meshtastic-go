package main

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
	"google.golang.org/protobuf/proto"
)

const start1 = byte(0x94)
const start2 = byte(0xc3)
const headerLen = 4
const maxToFromRadioSzie = 512
const broadcastAddr = "^all"
const localAddr = "^local"
const defaultHopLimit = 3
const broadcastNum = 0xffffffff

// Radio holds the port and serial io.ReadWriteCloser struct to maintain one serial connection
type Radio struct {
	portNumber string
	serialPort io.ReadWriteCloser
}

// Init initializes the Serial connection for the radio
func (r *Radio) Init(serialPort string) {
	r.portNumber = serialPort
	//Configure the serial port
	options := serial.OpenOptions{
		PortName:              r.portNumber,
		BaudRate:              921600,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       0,
		InterCharacterTimeout: 100,
		ParityMode:            serial.PARITY_NONE,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	r.serialPort = port
}

// sendPacket takes a protbuf packet, construct the appropriate header and sends it to the radio
func (r *Radio) sendPacket(protobufPacket []byte) (err error) {

	packageLength := len(string(protobufPacket))

	header := []byte{start1, start2, byte(packageLength>>8) & 0xff, byte(packageLength) & 0xff}

	radioPacket := append(header, protobufPacket...)
	_, err = r.serialPort.Write(radioPacket)
	if err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	return

}

// readResponse reads any responses in the serial port, convert them to a FromRadio protobuf and return
func (r *Radio) readResponse() (FromRadioPackets []*pb.FromRadio, err error) {

	b := make([]byte, 1)

	emptyByte := make([]byte, 0)
	processedBytes := make([]byte, 0)

	/************************************************************************************************
	* Process the returned data byte by byte until we have a valid command
	* Each command will come back with [START1, START2, PROTOBUF_PACKET]
	* where the protobuf packet is sent in binary. After reading START1 and START2
	* we use the next bytes to find the length of the packet.
	* After finding the length the looop continues to gather bytes until the length of the gathered
	* bytes is equal to the packet length plus the header
	 */
	for {

		_, err := r.serialPort.Read(b)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Reading Error: %v", err)
		}

		if len(b) > 0 {

			pointer := len(processedBytes)

			processedBytes = append(processedBytes, b...)

			if pointer == 0 {
				if b[0] != start1 {
					processedBytes = emptyByte
				}
			} else if pointer == 1 {
				if b[0] != start2 {
					processedBytes = emptyByte
				}
			} else if pointer >= headerLen {
				packetLength := int((processedBytes[2] << 8) + processedBytes[3])

				if pointer == headerLen {
					if packetLength > maxToFromRadioSzie {
						processedBytes = emptyByte
					}
				}

				if len(processedBytes) != 0 && pointer+1 == packetLength+headerLen {
					fromRadio := pb.FromRadio{}
					if err := proto.Unmarshal(processedBytes[headerLen:], &fromRadio); err != nil {
						return nil, err
					}
					FromRadioPackets = append(FromRadioPackets, &fromRadio)
					processedBytes = emptyByte
				}
			}

		} else {
			break
		}

	}

	return FromRadioPackets, nil

}

// GetRadioInfo retrieves information from the radio including config and adjacent Node information
func (r *Radio) GetRadioInfo() (radioResponses []*pb.FromRadio, err error) {
	// 42 seems to be the config for the CLI client.
	// TODO: Find if there's more significance or if it's just a hitchhikers refernce
	nodeInfo := pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: 42}}

	out, err := proto.Marshal(&nodeInfo)
	if err != nil {
		return nil, err
	}

	r.sendPacket(out)

	radioResponses, err = r.readResponse()

	return

}

// SendTextMessage sends a free form text message to other radios
func (r *Radio) SendTextMessage(message string, to int) error {
	var address int
	if to != 0 {
		address = to
	} else {
		address = broadcastNum
	}

	// This constant is defined in Constants_DATA_PAYLOAD_LEN, but not in a friendly way to use
	if len(message) > 240 {
		return errors.New("Message too large")
	}

	rand.Seed(time.Now().UnixNano())
	packetID := rand.Intn(2385286828-1) + 1

	radioMessage := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: &pb.MeshPacket{
				To:      uint32(address),
				WantAck: true,
				Id:      uint32(packetID),
				PayloadVariant: &pb.MeshPacket_Decoded{
					Decoded: &pb.SubPacket{
						PayloadVariant: &pb.SubPacket_Data{
							Data: &pb.Data{
								Payload: []byte(message),
								Portnum: pb.PortNum_TEXT_MESSAGE_APP,
							},
						},
					},
				},
			},
		},
	}

	out, err := proto.Marshal(&radioMessage)
	if err != nil {
		return err
	}

	if err := r.sendPacket(out); err != nil {
		return err
	}

	return nil

}

// SetRadioOwner sets the owner of the radio visible on the public mesh
func (r *Radio) SetRadioOwner(name string) error {

	if len(name) < 2 {
		return errors.New("Name too short")
	}

	packet := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_SetOwner{
			SetOwner: &pb.User{
				LongName:  name,
				ShortName: name[:3],
			},
		},
	}

	out, err := proto.Marshal(&packet)
	if err != nil {
		return err
	}

	if err := r.sendPacket(out); err != nil {
		return err
	}

	return nil
}

/*
   def channelURL(self):
       """The sharable URL that describes the current channel
       """
       bytes = self.radioConfig.channel_settings.SerializeToString()
       s = base64.urlsafe_b64encode(bytes).decode('ascii')
       return f"https://www.meshtastic.org/c/#{s}"

   def setURL(self, url, write=True):
       """Set mesh network URL"""
       if self.radioConfig == None:
           raise Exception("No RadioConfig has been read")

       # URLs are of the form https://www.meshtastic.org/c/#{base64_channel_settings}
       # Split on '/#' to find the base64 encoded channel settings
       splitURL = url.split("/#")
       decodedURL = base64.urlsafe_b64decode(splitURL[-1])
       self.radioConfig.channel_settings.ParseFromString(decodedURL)
       if write:
           self.writeConfig()
*/
// SetChannelURL sets the channel for the radio. The incoming channel should match the meshtastic URL format
// of a URL ending with /#{base_64_encoded_radio_params}
func (r *Radio) SetChannelURL(url string) error {

	split := strings.Split(url, "#")
	channel := split[len(split)-1]
	cDec, err := b64.StdEncoding.DecodeString(channel)
	if err != nil {
		return errors.New("Incorrect channel settings")
	}

	protoChannel := pb.RadioConfig{}

	if err := proto.Unmarshal(cDec, &protoChannel); err != nil {
		return err
	}

	fmt.Printf("Settings: %#v", &protoChannel)

	return nil
}

// Close closes the serial port. Added so users can defer the close after opening
func (r *Radio) Close() {
	r.serialPort.Close()
}
