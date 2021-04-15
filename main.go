package main

import (
	"fmt"
	"log"

	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
	"google.golang.org/protobuf/proto"
)

const BROADCAST_ADDR = "^all"
const LOCAL_ADDR = "^local"
const defaultHopLimit = 3
const BROADCAST_NUM = 0xffffffff

// option go_package = "github.com/lmatte7/meshtastic-go"

func main() {

	// node_info := &pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: 42}}
	radio_message := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: &pb.MeshPacket{
				To:      BROADCAST_NUM,
				WantAck: true,
				Id:      2338592482,
				PayloadVariant: &pb.MeshPacket_Decoded{
					Decoded: &pb.SubPacket{
						PayloadVariant: &pb.SubPacket_Data{
							Data: &pb.Data{
								Payload: []byte("This is a test message from Go!!!"),
								Portnum: pb.PortNum_TEXT_MESSAGE_APP,
							},
						},
					},
				},
			},
		},
	}

	fmt.Println(&radio_message)

	out, err := proto.Marshal(&radio_message)
	// out, err := proto.Marshal(node_info)
	if err != nil {
		log.Fatalln("Failed to encode protobuf:", err)
	}

	// fmt.Printf("Encoded Buff: %q", out)

	radio := Radio{port_number: "/dev/cu.SLAB_USBtoUART"}

	radio.Init()

	defer radio.Close()

	radio.SendPacket(out)

	responses := radio.ReadResponse()

	for _, response := range responses {
		fmt.Println(response)
	}

}
