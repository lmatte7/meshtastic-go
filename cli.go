package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
)

//TODO: Adjust command structure to match the below

/*

Commands
port
recv
sendText

Subcommands

Info
info - show all info
info:channels
info:nodes - show nodes


messages - display any stored messages
messages:wait - wait for new messages

Position
position - show position
position:set:lat LAT
position:set:long LONG
position:set:alt ALT

Channel
channel - show channel information
channel:add NAME INDEX
channel:delete INDEX
channel:set INDEX KEY VALUE

Preference
preference - show preferences
preference:set KEY VALUE
preferences:setowner OWNER
*/

// Init starts the CLI and determines flags
func Init() {

	app := &cli.App{
		Name:    "meshtastic-go",
		Version: "v0.2",
		Authors: []*cli.Author{
			{
				Name:  "Lucas Matte",
				Email: "lmatte7@gmail.com",
			},
		},
		Usage: "Interface with meshtastic radios",
		Commands: []*cli.Command{
			{
				Name:        "info",
				Usage:       "Show radio information",
				UsageText:   "info [command] - Show radio information",
				Description: "Show node, preference and channel information for radio",
				ArgsUsage:   "",
				Action:      showAllRadioInfo,
				Subcommands: []*cli.Command{
					{
						Name:    "radio",
						Aliases: []string{"r"},
						Usage:   "Show radio information",
						Action:  showRadioInfo,
					},
					{
						Name:    "channels",
						Aliases: []string{"c"},
						Usage:   "Show all channel information",
						Action:  showChannelInfo,
					},
					{
						Name:    "nodes",
						Aliases: []string{"n"},
						Usage:   "Show all nodes on the mesh",
						Action:  showNodeInfo,
					},
					{
						Name:    "preferences",
						Aliases: []string{"p"},
						Usage:   "Show radio user preferences",
						Action:  showRadioPreferences,
					},
				},
			},
			{
				Name:        "message",
				Usage:       "Show radio information",
				UsageText:   "info [command] - Show radio information",
				Description: "Show node, preference and channel information for radio",
				ArgsUsage:   "",
				Action:      showAllRadioInfo,
				Subcommands: []*cli.Command{
					{
						Name:        "send",
						Usage:       "Send a text message",
						UsageText:   "text - Sends a text message to a node",
						Description: "Sends a text message to a Node, or to all nodes if no address is provided",
						Action:      sendText,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "message",
								Aliases:  []string{"m"},
								Usage:    "Message to send, must be under 240 characters",
								Required: true,
							},
							&cli.Int64Flag{
								Name:    "to",
								Aliases: []string{"t"},
								Usage:   "Address to send to. Leave blank for broadcast",
								Value:   0,
							},
						},
					},
					{
						Name:        "recv",
						Usage:       "Wait for new messages",
						Description: "Waits for new messages and displays them as recieved until cancelled. Only shows messages on TEXT_MESSAGE port",
						Action:      getRecievedMessages,
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Usage:       "specify a port",
				Value:       "/dev/cu.SLAB_USBtoUART",
				DefaultText: "/dev/cu.SLAB_USBtoUART",
			},
		},
	}

	app.Run(os.Args)

}

func showAllRadioInfo(c *cli.Context) error {

	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = getRadioInfo(radio)
	if err != nil {
		return err
	}

	return nil

}

func showNodeInfo(c *cli.Context) error {

	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = displayNodes(radio)
	if err != nil {
		return err
	}

	return nil

}

func showRadioInfo(c *cli.Context) error {

	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	responses, err := radio.GetRadioInfo()
	if err != nil {
		return err
	}

	for _, response := range responses {

		if info, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_MyInfo); ok {
			printRadioInfo(info)
		}
	}

	return nil

}

func showRadioPreferences(c *cli.Context) error {
	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = printRadioPreferences(radio)
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
