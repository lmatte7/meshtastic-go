package main

import (
	"fmt"
	"io"
	"log"
	"strconv"

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

	nodeInfo := &pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: 42}}

	fmt.Println(nodeInfo)

	out, err := proto.Marshal(nodeInfo)
	if err != nil {
		log.Fatalln("Failed to encode address book:", err)
	}
	options := serial.OpenOptions{
		PortName:        "/dev/cu.SLAB_USBtoUART",
		BaudRate:        921600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// fmt.Printf("%q\n", out)

	// Make sure to close it later.
	defer port.Close()

	package_length := len(out)

	header := []byte{START1, START2, byte(package_length>>8) & 0xff, byte(package_length)}
	// header = bytes([START1, START2, (bufLen >> 8) & 0xff,  bufLen & 0xff])

	fmt.Println(package_length)
	fmt.Printf("convert to byte: %q\n", byte(package_length>>16))
	fmt.Printf("%q\n", header)
	fmt.Printf("%q\n", out)

	header = append(header, out...)

	fmt.Printf("final sequence: %q\n", header)

	_, err = port.Write(header)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}

	b := make([]byte, 32)

	for {
		_, err := port.Read(b)
		if err != nil {
			log.Fatalf("port.Write: %v", err)
		}
		// fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
		// fmt.Printf("%q\n", b[:n])
		fmt.Println(strconv.Unquote(string(b)))
		if err == io.EOF {
			break
		}
	}

}
