package main

import (
	"io"
	"log"
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
func (radio *Radio) Init() {
	//Configure the serial port
	/*
		TODO: Come up with a way to detect the end of the stream
		* The EOF error comes up and that ends the loop, but it'd be better
		* to not have the for loop break on a error */
	options := serial.OpenOptions{
		PortName:              radio.portNumber,
		BaudRate:              921600,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       0,
		InterCharacterTimeout: 500,
		ParityMode:            serial.PARITY_NONE,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	radio.serialPort = port
}

// sendPacket takes a protbuf packet, construct the appropriate header and sends it to the radio
func (radio *Radio) sendPacket(protobufPacket []byte) (err error) {

	packageLength := len(string(protobufPacket))

	header := []byte{start1, start2, byte(packageLength>>8) & 0xff, byte(packageLength) & 0xff}

	radioPacket := append(header, protobufPacket...)
	_, err = radio.serialPort.Write(radioPacket)
	if err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	return

}

// readResponse reads any responses in the serial port, convert them to a FromRadio protobuf and return
func (radio *Radio) readResponse() (FromRadioPackets []*pb.FromRadio, err error) {

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

		_, err := radio.serialPort.Read(b)
		if err != nil {
			break
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
func (radio *Radio) GetRadioInfo() (radioResponses []*pb.FromRadio, err error) {
	// 42 seems to be the config for the CLI client.
	// TODO: Find if there's more significance or if it's just a hitchhikers refernce
	nodeInfo := pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: 42}}

	out, err := proto.Marshal(&nodeInfo)
	if err != nil {
		return nil, err
	}

	radio.sendPacket(out)

	radioResponses, err = radio.readResponse()

	return

}

// SendTextMessage sends a free form text message to other radios
// TODO: Add limit for string
func (radio *Radio) SendTextMessage(message string) error {
	// node_info := &pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: 42}}
	radioMessage := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: &pb.MeshPacket{
				To:      broadcastNum,
				WantAck: true,
				Id:      2338592482,
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

	if err := radio.sendPacket(out); err != nil {
		return err
	}

	return nil

}

// Close closes the serial port. Added so users can defer the close after opening
func (radio *Radio) Close() {
	radio.serialPort.Close()
}
