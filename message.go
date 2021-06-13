package main

import (
	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
)

func getRecievedMessages(c *cli.Context) error {

	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	printMessageHeader()
	for {

		responses, err := radio.GetRadioInfo()
		if err != nil {
			return err
		}

		recievedMessages := make([]*gomeshproto.FromRadio_Packet, 0)

		for _, response := range responses {
			if packet, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_Packet); ok {
				if packet.Packet.GetDecoded().GetPortnum() == gomeshproto.PortNum_TEXT_MESSAGE_APP {
					recievedMessages = append(recievedMessages, packet)
				}
			}
		}

		if len(recievedMessages) > 0 {
			printMessages(recievedMessages)
		}
	}

}

func sendText(c *cli.Context) error {

	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = radio.SendTextMessage(c.String("message"), c.Int64("to"))
	if err != nil {
		return err
	}

	return nil
}
