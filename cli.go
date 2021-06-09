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
info:nodes - show nodes

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
				Name:        "text",
				Usage:       "Send a text message",
				UsageText:   "text - Sends a text message to a node",
				Description: "Sends a text message to a Node, or to all nodes if no address is provided",
				ArgsUsage:   "[text to]",
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
				Name:        "info",
				Usage:       "Show radio information",
				UsageText:   "info [command] - Show radio information",
				Description: "Show node, preference and channel information for radio",
				ArgsUsage:   "",
				Subcommands: []*cli.Command{
					{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Show all radio information",
						Action:  showAllRadioInfo,
					},
					{
						Name:    "channels",
						Aliases: []string{"c"},
						Usage:   "Show all channel information",
						Action:  displayChannelInfo,
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

	// fmt.Println(app)

	// var port string
	// var message string
	// var owner string
	// var setKey string
	// var setValue string
	// var url string
	// var to int64

	// flag.StringVar(&port, "port", "", "--port=port The serial port for the radio (Required)")
	// flag.StringVar(&message, "text", "", "--text=message Send a text message")
	// flag.StringVar(&url, "url", "", "--url=channel_curl Set the radio URL")
	// flag.StringVar(&setKey, "setKey", "", "--setKey=key The key to set for a custom user preference option. Used with setValue")
	// flag.StringVar(&setValue, "setValue", "", "--setValue=value The value to set for a custom user preference option. Used with setKey")
	// flag.StringVar(&owner, "setOwner", "", "--setowner=owner Set the listed owner for the radio")
	// flag.Int64Var(&to, "to", 0, "--to=destination Node to receive text (Used with sendtext)")
	// recv := flag.Bool("recv", false, "Wait for new messages")
	// infoPtr := flag.Bool("info", false, "Display radio information")
	// longslowPtr := flag.Bool("longSlow", false, "Set long-range but slow channel")
	// shortFast := flag.Bool("shortFast", false, "Set short-range but fast channel")

	// flag.Usage = func() {
	// 	flagSet := flag.CommandLine
	// 	fmt.Printf("A command line tool for interacting with meshtastic radios\n")
	// 	fmt.Printf("\n")
	// 	fmt.Printf("USAGE\n")
	// 	fmt.Printf("meshtastic-go -p <port> [COMMAND]\n")
	// 	fmt.Printf("\n")
	// 	fmt.Printf("COMMANDS\n")
	// 	fmt.Printf("\n")
	// 	order := []string{"port", "info", "recv", "longSlow", "shortFast", "text", "setOwner", "setKey", "setValue"}
	// 	for _, name := range order {
	// 		flag := flagSet.Lookup(name)
	// 		fmt.Printf("--%-15s", flag.Name)
	// 		fmt.Printf("%-10s\n", flag.Usage)
	// 	}
	// }

	// flag.Parse()

	// if port == "" {
	// 	flag.Usage()
	// 	os.Exit(1)
	// }

	// radio := gomesh.Radio{}
	// if len(message) > 0 || *infoPtr || *recv || len(owner) > 0 || len(url) > 0 || *longslowPtr || *shortFast || len(setKey) > 0 || len(setValue) > 0 {
	// 	radio.Init(port)
	// 	defer radio.Close()
	// }

	// if message != "" {
	// 	err := radio.SendTextMessage(message, to)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	}
	// }

	// if *recv {
	// 	err := getRecievedMessages(radio)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	}
	// }

	// if *infoPtr {
	// 	err := getRadioInfo(radio)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	}
	// }

	// if *longslowPtr {
	// 	err := radio.SetChannel(0)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	} else {
	// 		fmt.Printf("Set channel\n")
	// 	}
	// }

	// if *shortFast {
	// 	err := radio.SetChannel(1)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	} else {
	// 		fmt.Printf("Set channel\n")
	// 	}
	// }

	// if url != "" {
	// 	err := radio.SetChannelURL(url)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	} else {
	// 		fmt.Printf("Set URL successful\n")
	// 	}
	// }

	// if owner != "" {
	// 	err := radio.SetRadioOwner(owner)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	} else {
	// 		fmt.Printf("Set owner successful\n")
	// 	}
	// }

	// if setValue != "" && setKey != "" {
	// 	err := radio.SetUserPreferences(setKey, setValue)
	// 	if err != nil {
	// 		fmt.Printf("ERROR: %s\n", err)
	// 	} else {
	// 		fmt.Printf("Set user preferences successful\n")
	// 	}
	// }

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

func displayChannelInfo(c *cli.Context) error {
	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = printChannelSettings(radio)
	if err != nil {
		return err
	}

	return nil
}

// func getRecievedMessages(r gomesh.Radio) error {

// 	printMessageHeader()
// 	for {

// 		responses, err := r.GetRadioInfo()
// 		if err != nil {
// 			return err
// 		}

// 		recievedMessages := make([]*gomeshproto.FromRadio_Packet, 0)

// 		for _, response := range responses {
// 			if packet, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_Packet); ok {
// 				recievedMessages = append(recievedMessages, packet)
// 			}
// 		}

// 		if len(recievedMessages) > 0 {
// 			printMessages(recievedMessages)
// 		}
// 	}

// }

func getRadioInfo(r gomesh.Radio) error {

	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	nodes := make([]*gomeshproto.FromRadio_NodeInfo, 0)
	recievedMessages := make([]*gomeshproto.FromRadio_Packet, 0)

	for _, response := range responses {

		// fmt.Printf("Responses: %v\n\n", response)
		if info, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_MyInfo); ok {
			fmt.Printf("%s", "\nRadio Settings: \n")
			fmt.Printf("%-25s", "Node Number: ")
			fmt.Printf("%d\n", info.MyInfo.MyNodeNum)
			fmt.Printf("%-25s", "GPS: ")
			fmt.Printf("%t\n", info.MyInfo.HasGps)
			fmt.Printf("%-25s", "Number of Channels: ")
			fmt.Printf("%d\n", info.MyInfo.MaxChannels)
			fmt.Printf("%-25s", "Firmware: ")
			fmt.Printf("%s\n", info.MyInfo.FirmwareVersion)
			fmt.Printf("%-25s", "Message Timeout (msec): ")
			fmt.Printf("%d\n", info.MyInfo.MessageTimeoutMsec)
			fmt.Printf("%-25s", "Min App Version: ")
			fmt.Printf("%d\n", info.MyInfo.MinAppVersion)

		}

		// if radioInfo, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_Radio); ok {
		// 	if radioInfo.Radio.Preferences != nil {
		// 		fmt.Printf("\n")
		// 		fmt.Printf("Preferences:\n")
		// 		fmt.Printf("%-25s", "ls secs: ")
		// 		fmt.Printf("%d\n", radioInfo.Radio.Preferences.LsSecs)
		// 		fmt.Printf("%-25s", "Region: ")
		// 		fmt.Printf("%d\n", radioInfo.Radio.Preferences.Region)
		// 	}

		// 	if radioInfo.Radio.ChannelSettings != nil {
		// 		fmt.Printf("\n")
		// 		fmt.Printf("Channel Settings:\n")
		// 		fmt.Printf("%-25s", "Modem Config: ")
		// 		fmt.Printf("%s\n", radioInfo.Radio.ChannelSettings.ModemConfig)
		// 		fmt.Printf("%-25s", "PSK: ")
		// 		fmt.Printf("%q\n", radioInfo.Radio.ChannelSettings.Psk)

		// 		protoChannel := radioInfo.Radio.ChannelSettings

		// 		out, err := proto.Marshal(protoChannel)
		// 		if err != nil {
		// 			fmt.Printf("ERROR: Error parsing channel URL")
		// 		}

		// 		url := b64.StdEncoding.EncodeToString(out)

		// 		fmt.Printf("%-25s", "Channel URL: ")
		// 		fmt.Printf("https://www.meshtastic.org/c/#%s\n", url)
		// 	}
		// }

		if nodeInfo, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}

		if packet, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_Packet); ok {
			if packet.Packet.GetDecoded().Portnum == gomeshproto.PortNum_TEXT_MESSAGE_APP {
				recievedMessages = append(recievedMessages, packet)
			}
		}

	}

	err = printChannelSettings(r)
	if err != nil {
		return err
	}

	if len(nodes) > 0 {
		fmt.Printf("\n")
		fmt.Printf("Nodes in Mesh:\n")

		fmt.Printf("%-80s", "======================================================================================\n")
		fmt.Printf("| %-15s| ", "Node Number")
		fmt.Printf("%-15s| ", "User")
		fmt.Printf("%-15s| ", "Battery")
		fmt.Printf("%-15s| ", "Latitude")
		fmt.Printf("%-15s", "Longitude      |\n")
		fmt.Printf("%-80s", "--------------------------------------------------------------------------------------\n")
		for _, node := range nodes {
			if node.NodeInfo != nil {
				fmt.Printf("| %-15s| ", fmt.Sprint(node.NodeInfo.Num))
				if node.NodeInfo.User != nil {
					fmt.Printf("%-15s| ", node.NodeInfo.User.LongName)
				} else {
					fmt.Printf("%-15s| ", "N/A")
				}
				if node.NodeInfo.Position != nil {
					fmt.Printf("%-15s| ", fmt.Sprint(node.NodeInfo.Position.BatteryLevel))
					fmt.Printf("%-15s| ", fmt.Sprint(node.NodeInfo.Position.LatitudeI))
					fmt.Printf("%-15s", fmt.Sprint(node.NodeInfo.Position.LongitudeI))
				} else {
					fmt.Printf("%-15s| ", "N/A")
					fmt.Printf("%-15s| ", "N/A")
					fmt.Printf("%-15s| ", "N/A")
				}
				fmt.Printf("%s", "|\n")
			}
		}
		fmt.Printf("%-80s", "======================================================================================\n")
	}

	if len(recievedMessages) > 0 {
		printMessageHeader()
		printMessages(recievedMessages)
		fmt.Printf("%-80s", "--------------------------------------------------------------------------------------\n")
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
		// fmt.Printf("%-18s| ", message.Packet.GetDecoded().GetData().GetPortnum())
		// re := regexp.MustCompile(`\r?\n`)
		// escMesg := re.ReplaceAllString(string(message.Packet.GetDecoded().GetData().Payload), "")
		// fmt.Printf("%-53q", escMesg)
		fmt.Printf("%s", "|\n")
	}
}

func printChannelSettings(r gomesh.Radio) error {

	channels := []gomeshproto.AdminMessage{}
	channelCount := 0
	for {

		info, err := r.GetChannelInfo(channelCount)
		if err != nil {
			return err
		}
		if info.GetGetChannelResponse().Role == gomeshproto.Channel_DISABLED {
			break
		}

		channels = append(channels, info)
		// Add a guarenteed exit for the loop since there can't be more than 20 channels
		channelCount++
		if channelCount > 20 {
			break
		}
	}

	fmt.Printf("%s", "\n")
	fmt.Printf("Channel Settings:\n")
	fmt.Printf("%-80s", "================================================================================================================================================\n")
	fmt.Printf("| %-15s| ", "Name")
	fmt.Printf("%-15s| ", "Role")
	fmt.Printf("%-15s| ", "Modem")
	fmt.Printf("%-90s", "PSK")
	fmt.Printf("%s", "|\n")
	fmt.Printf("%-80s", "------------------------------------------------------------------------------------------------------------------------------------------------\n")
	for _, channelInfo := range channels {

		if channelInfo.GetGetChannelResponse().Role == gomeshproto.Channel_DISABLED {
			break
		}
		if len(channelInfo.GetGetChannelResponse().GetSettings().Name) > 0 {
			fmt.Printf("| %-15s| ", channelInfo.GetGetChannelResponse().GetSettings().Name)
		} else {
			fmt.Printf("| %-15s| ", "N/A")
		}
		if len(channelInfo.GetGetChannelResponse().Role.String()) > 0 {
			fmt.Printf("%-15s| ", channelInfo.GetGetChannelResponse().Role.String())
		} else {
			fmt.Printf("%-15s| ", "N/A")
		}
		if len(channelInfo.GetGetChannelResponse().GetSettings().ModemConfig.String()) > 0 {
			fmt.Printf("%-15s| ", channelInfo.GetGetChannelResponse().GetSettings().ModemConfig)
		} else {
			fmt.Printf("%-15s| ", "N/A")
		}
		if len(channelInfo.GetGetChannelResponse().GetSettings().Psk) > 0 {
			re := regexp.MustCompile(`\r?\n`)
			escMesg := re.ReplaceAllString(string(channelInfo.GetGetChannelResponse().GetSettings().Psk), "")
			fmt.Printf("%-90q", escMesg)
		} else {
			fmt.Printf("%-53s| ", "N/A")
		}
		fmt.Printf("%s", "|\n")

	}
	fmt.Printf("%-80s", "================================================================================================================================================\n")

	return nil
}
