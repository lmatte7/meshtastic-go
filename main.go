package main

import (
	"fmt"
	"log"

	"github.com/jacobsa/go-serial/serial"
	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
	"google.golang.org/protobuf/proto"
)

func main() {

	//port := "/dev/cu.SLAB_USBtoUART"

	// config := &pb.AdminMessage{}

	// config.GetGetRadioRequest()

	START1 := byte(0x94)
	START2 := byte(0xc3)
	HEADER_LEN := 4
	MAX_TO_FROM_RADIO_SIZE := 512

	nodeInfo := &pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: 42}}

	out, err := proto.Marshal(nodeInfo)
	if err != nil {
		log.Fatalln("Failed to encode address book:", err)
	}
	options := serial.OpenOptions{
		PortName:              "/dev/cu.SLAB_USBtoUART",
		BaudRate:              921600,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       0,
		InterCharacterTimeout: 500,
		ParityMode:            serial.PARITY_NONE,
	}

	package_length := len(out)

	// wake_up := []byte{START1, START1, START1, START1}

	header := []byte{START1, START2, byte(package_length>>8) & 0xff, byte(package_length)}

	header = append(header, out...)

	fmt.Printf("final sequence: %q\n", header)

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// fmt.Printf("%q\n", out)

	// Make sure to close it later.
	defer port.Close()

	// _, err = port.Write(wake_up)
	// _, err = port.Read(b)
	// if err != nil {
	// 	log.Fatalf("port.Write: %v", err)
	// }
	_, err = port.Write(header)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}

	fromRadio := &pb.FromRadio{}

	b := make([]byte, 1)

	emptyByte := make([]byte, 0)
	processedBytes := make([]byte, 0)

	for {

		_, err = port.Read(b)
		if err != nil {
			log.Fatalf("port.Write read Error: %v", err)
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
					if err := proto.Unmarshal(processedBytes[HEADER_LEN:], fromRadio); err != nil {
						log.Fatalln("Failed to parse packet:", err)
					}
					fmt.Println(fromRadio)
					fmt.Println("")
					processedBytes = emptyByte
				}
			}

		} else {
			break
		}
	}

	// for {
	// 	_, err := port.Read(b)
	// 	if err != nil {
	// 		log.Fatalf("port.Write read Error: %v", err)
	// 	}
	// 	// for len(b) > 0 {
	// 	// 	r, size := utf8.DecodeRune(b)
	// 	// 	fmt.Printf("%c", r)

	// 	// 	b = b[size:]
	// 	// }
	// 	// fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
	// 	// fmt.Println(string(b))
	// 	// for _, n := range b[18:] {
	// 	// 	fmt.Printf("% 08b", n) // prints 00000000 11111101
	// 	// }
	// 	fmt.Println(b[23:])
	// 	fmt.Printf("%q\n", b[23:])
	// 	if err := proto.Unmarshal(b[18:], fromRadio); err != nil {
	// 		log.Fatalln("Failed to parse packet:", err)
	// 	}
	// 	fmt.Println(fromRadio)
	// 	if err == io.EOF {
	// 		break
	// 	}
	// }

}
