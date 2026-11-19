package gmtcli

import (
	"fmt"

	"github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
)

// rebootRadio is the action for the reboot command.
func rebootRadio(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	myNode, _, err := radioNodes(radio)
	if err != nil {
		return cli.Exit(err, 1)
	}

	packet, err := rebootPacket(myNode, 1)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if err := radio.SendPacket(packet); err != nil {
		return cli.Exit(fmt.Sprintf("sending reboot: %v", err), 1)
	}

	fmt.Println("Reboot requested; the radio will be back in a few seconds.")
	return nil
}

// rebootPacket builds the ToRadio message carrying an admin reboot request,
// addressed to the radio itself, that fires after the given delay.
func rebootPacket(myNode uint32, afterSeconds int32) ([]byte, error) {
	admin, err := proto.Marshal(&gomeshproto.AdminMessage{
		PayloadVariant: &gomeshproto.AdminMessage_RebootSeconds{RebootSeconds: afterSeconds},
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
