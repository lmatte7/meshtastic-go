package main

import (
	"fmt"
	"regexp"

	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
)

func getReceivedMessages(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	if !c.Bool("json") {
		printMessageHeader()
	}
	for {

		responses, err := radio.ReadResponse(false)
		if err != nil {
			return cli.Exit(err.Error(), 0)
		}

		receivedMessages := []*gomeshproto.FromRadio_Packet{}

		for _, response := range responses {
			if packet, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_Packet); ok {
				if packet.Packet.GetDecoded().GetPortnum() == gomeshproto.PortNum_TEXT_MESSAGE_APP {
					receivedMessages = append(receivedMessages, packet)
				}
			}
		}

		if len(receivedMessages) > 0 {
			if c.Bool("json") {
				printJsonMessages(receivedMessages)
			} else {
				printMessages(receivedMessages)
			}
			if c.Bool("exit") {
				return nil
			}
		}
	}

}

func sendText(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	err := radio.SendTextMessage(c.String("message"), c.Int64("to"), c.Int64("channel"))
	if err != nil {
		return cli.Exit(err, 0)
	}

	return nil
}

func printMessageHeader() {
	fmt.Printf("\n")
	fmt.Printf("Received Messages:\n")
	printDoubleDivider()
	fmt.Printf("| %-15s| ", "From")
	fmt.Printf("%-15s| ", "To")
	fmt.Printf("%-18s| ", "Port Num")
	fmt.Printf("%-10s| ", "Channel")
	fmt.Printf("%-15s ", "Payload                                              |\n")
	printSingleDivider()
}

func printMessages(messages []*gomeshproto.FromRadio_Packet) {
	for _, message := range messages {
		fmt.Printf("| %-15s| ", fmt.Sprint(message.Packet.From))
		fmt.Printf("%-15s| ", fmt.Sprint(message.Packet.To))
		fmt.Printf("%-18s| ", message.Packet.GetDecoded().GetPortnum().String())
		fmt.Printf("%-10s| ", fmt.Sprint(message.Packet.Channel))
		re := regexp.MustCompile(`\r?\n`)
		escMesg := re.ReplaceAllString(string(message.Packet.GetDecoded().Payload), "")
		fmt.Printf("%-53q", escMesg)
		fmt.Printf("%s", "|\n")
	}
}

func printJsonMessages(messages []*gomeshproto.FromRadio_Packet) {
	for _, message := range messages {
		fmt.Printf("{\"from\":%s,", fmt.Sprint(message.Packet.From))
		fmt.Printf("\"to\":%s,", fmt.Sprint(message.Packet.To))
		fmt.Printf("\"portnum\": \"%s\",", message.Packet.GetDecoded().GetPortnum().String())
		fmt.Printf("\"channel\":%s,", fmt.Sprint(message.Packet.Channel))
		re := regexp.MustCompile(`\r?\n`)
		escMesg := re.ReplaceAllString(string(message.Packet.GetDecoded().Payload), "")
		fmt.Printf("\"Payload\": \"%s\"", escMesg)
		fmt.Printf("%s", "}\n")
	}
}
