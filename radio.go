package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jacobsa/go-serial/serial"
	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
	"google.golang.org/protobuf/proto"
)

const START1 = byte(0x94)
const START2 = byte(0xc3)
const HEADER_LEN = 4
const MAX_TO_FROM_RADIO_SIZE = 512

type Radio struct {
	port_number string
	SerialPort  io.ReadWriteCloser
}

func (radio *Radio) Init() {
	//Configure the serial port
	// TODO: Come up with a way to detect the end of the stream
	// The EOF error comes up and that ends the loop, but it'd be better
	// to not have the for loop break on a error
	options := serial.OpenOptions{
		PortName:              radio.port_number,
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

	radio.SerialPort = port
}

func (radio *Radio) SendPacket(protobuf_packet []byte) {

	package_length := len(protobuf_packet)

	header := []byte{START1, START2, byte(package_length>>8) & 0xff, byte(package_length)}

	radio_packet := append(header, protobuf_packet...)
	_, err := radio.SerialPort.Write(radio_packet)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

}

func (radio *Radio) ReadResponse() (FromRadioPackets []*pb.FromRadio) {

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

		_, err := radio.SerialPort.Read(b)
		if err != nil {
			break
		}

		if len(b) > 0 {

			pointer := len(processedBytes)

			processedBytes = append(processedBytes, b...)

			if pointer == 0 {
				if b[0] != START1 {
					processedBytes = emptyByte
				}
			} else if pointer == 1 {
				if b[0] != START2 {
					processedBytes = emptyByte
				}
			} else if pointer >= HEADER_LEN {
				packet_length := int((processedBytes[2] << 8) + processedBytes[3])

				if pointer == HEADER_LEN {
					if packet_length > MAX_TO_FROM_RADIO_SIZE {
						processedBytes = emptyByte
						fmt.Println("Start over")
					}
				}

				if len(processedBytes) != 0 && pointer+1 == packet_length+HEADER_LEN {
					fromRadio := pb.FromRadio{}
					if err := proto.Unmarshal(processedBytes[HEADER_LEN:], &fromRadio); err != nil {
						log.Fatalln("Failed to parse packet:", err)
					}
					FromRadioPackets = append(FromRadioPackets, &fromRadio)
					processedBytes = emptyByte
				}
			}

		} else {
			break
		}

	}

	return FromRadioPackets

}

func (radio *Radio) Close() {
	radio.SerialPort.Close()
}
