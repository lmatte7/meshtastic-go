package main

import (
	"fmt"
	"regexp"

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

		responses, err := radio.ReadResponse()
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
			if c.Bool("exit") {
				return nil
			}
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

func printMessageHeader() {
	fmt.Printf("\n")
	fmt.Printf("Recieved Messages:\n")
	fmt.Printf("%-80s", "==============================================================================================================\n")
	fmt.Printf("| %-15s| ", "From")
	fmt.Printf("%-15s| ", "To")
	fmt.Printf("%-18s| ", "Port Num")
	fmt.Printf("%-15s ", "Payload                                              |\n")
	fmt.Printf("%-80s", "-------------------------------------------------------------------------------------------------------------\n")
}

func printMessages(messages []*gomeshproto.FromRadio_Packet) {

	for _, message := range messages {
		fmt.Printf("| %-15s| ", fmt.Sprint(message.Packet.From))
		fmt.Printf("%-15s| ", fmt.Sprint(message.Packet.To))
		fmt.Printf("%-18s| ", message.Packet.GetDecoded().GetPortnum().String())
		re := regexp.MustCompile(`\r?\n`)
		escMesg := re.ReplaceAllString(string(message.Packet.GetDecoded().Payload), "")
		fmt.Printf("%-53q", escMesg)
		fmt.Printf("%s", "|\n")
	}
}
