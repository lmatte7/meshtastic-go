package main

import (
	"fmt"
	"log"

	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
	"google.golang.org/protobuf/proto"
)

func main() {

	nodeInfo := &pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: 42}}

	out, err := proto.Marshal(nodeInfo)
	if err != nil {
		log.Fatalln("Failed to encode address book:", err)
	}

	radio := Radio{port_number: "/dev/cu.SLAB_USBtoUART"}

	radio.Init()

	defer radio.Close()

	radio.SendPacket(out)

	responses := radio.ReadResponse()

	for _, response := range responses {
		fmt.Println(response)
	}

}
