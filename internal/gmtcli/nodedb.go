package gmtcli

import (
	"errors"
	"fmt"

	"github.com/lmatte7/meshtastic-go/pkg/gomesh"
	"github.com/lmatte7/meshtastic-go/pkg/gomesh/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
)

// radioNodes asks the radio for its database and returns its own node number
// and every node it knows about.
func radioNodes(radio gomesh.Radio) (myNode uint32, nodes []MeshNode, err error) {
	responses, err := radio.GetRadioInfo()
	if err != nil {
		return 0, nil, err
	}
	for _, response := range responses {
		switch payload := response.GetPayloadVariant().(type) {
		case *gomeshproto.FromRadio_MyInfo:
			myNode = payload.MyInfo.GetMyNodeNum()
		case *gomeshproto.FromRadio_NodeInfo:
			nodes = append(nodes, ProtoNodeToMeshNode(payload))
		}
	}
	if myNode == 0 {
		return 0, nil, errors.New("radio did not report its node number")
	}
	return myNode, nodes, nil
}

// resetNodeDB is the action for "nodedb reset": it asks the radio to forget
// every node it knows about, after which it rebuilds the table from the
// nodes it hears.
func resetNodeDB(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	myNode, nodes, err := radioNodes(radio)
	if err != nil {
		return cli.Exit(err, 1)
	}

	packet, err := nodeDBResetPacket(myNode)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if err := radio.SendPacket(packet); err != nil {
		return cli.Exit(fmt.Sprintf("sending node database reset: %v", err), 1)
	}

	others := len(nodes) - 1 // the radio's own entry survives
	if others < 0 {
		others = 0
	}
	fmt.Printf("Node database reset requested; the radio will forget its %d other nodes and rebuild the table from what it hears.\n", others)
	return nil
}

// nodeDBResetPacket builds the ToRadio message carrying an admin
// nodedb_reset request addressed to the radio itself.
func nodeDBResetPacket(myNode uint32) ([]byte, error) {
	admin, err := proto.Marshal(&gomeshproto.AdminMessage{
		PayloadVariant: &gomeshproto.AdminMessage_NodedbReset{NodedbReset: true},
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&gomeshproto.ToRadio{PayloadVariant: &gomeshproto.ToRadio_Packet{Packet: &gomeshproto.MeshPacket{
		To:      myNode,
		WantAck: true,
		PayloadVariant: &gomeshproto.MeshPacket_Decoded{Decoded: &gomeshproto.Data{
			Portnum:      gomeshproto.PortNum_ADMIN_APP,
			Payload:      admin,
			WantResponse: true,
		}},
	}}})
}
