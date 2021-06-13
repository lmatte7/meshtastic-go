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

func displayNodes(r gomesh.Radio) error {
	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	nodes := make([]*gomeshproto.FromRadio_NodeInfo, 0)
	for _, response := range responses {
		if nodeInfo, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}
	}

	printNodes(nodes)

	return nil
}

func getRadioInfo(r gomesh.Radio) error {

	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	nodes := make([]*gomeshproto.FromRadio_NodeInfo, 0)
	recievedMessages := make([]*gomeshproto.FromRadio_Packet, 0)

	for _, response := range responses {

		if info, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_MyInfo); ok {
			printRadioInfo(info)
		}

		if nodeInfo, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}

		if packet, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_Packet); ok {
			if packet.Packet.GetDecoded().Portnum == gomeshproto.PortNum_TEXT_MESSAGE_APP {
				recievedMessages = append(recievedMessages, packet)
			}
		}

	}

	err = printRadioPreferences(r)
	if err != nil {
		return err
	}

	err = printChannelSettings(r)
	if err != nil {
		return err
	}

	if len(nodes) > 0 {
		printNodes(nodes)
	}

	if len(recievedMessages) > 0 {
		printMessageHeader()
		printMessages(recievedMessages)
		fmt.Printf("%-80s", "--------------------------------------------------------------------------------------\n")
	}

	return nil
}

func printNodes(nodes []*gomeshproto.FromRadio_NodeInfo) {
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

func printRadioInfo(info *gomeshproto.FromRadio_MyInfo) {
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

func printRadioPreferences(r gomesh.Radio) error {

	prefs, err := r.GetRadioPreferences()
	if err != nil {
		return err
	}

	fmt.Printf("%s", "\n")
	fmt.Printf("Radio Preferences:\n")

	fmt.Printf("%-25s", "Position Broadcast Secs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().PositionBroadcastSecs)
	fmt.Printf("%-25s", "Send Owner Interval:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().SendOwnerInterval)
	fmt.Printf("%-25s", "Wait Bluetooth (secs):")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().WaitBluetoothSecs)
	fmt.Printf("%-25s", "Screen On (secs):")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().ScreenOnSecs)
	fmt.Printf("%-25s", "Phone Timeout (secs):")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().PhoneTimeoutSecs)
	fmt.Printf("%-25s", "Phone Sds Timeout (secs):")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().PhoneSdsTimeoutSec)
	fmt.Printf("%-25s", "Mesh Sds Timeout (secs):")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().MeshSdsTimeoutSecs)
	fmt.Printf("%-25s", "Sds Secs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().SdsSecs)
	fmt.Printf("%-25s", "Ls Secs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().LsSecs)
	fmt.Printf("%-25s", "Min Wake (secs):")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().MinWakeSecs)

	if len(prefs.GetGetRadioResponse().GetPreferences().WifiSsid) > 0 {
		fmt.Printf("%-25s", "Wifi SSID:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().WifiSsid)
	} else {
		fmt.Printf("%-25s", "Wifi SSID:")
		fmt.Printf("%s\n", "N/A")
	}
	if len(prefs.GetGetRadioResponse().GetPreferences().WifiPassword) > 0 {
		fmt.Printf("%-25s", "Wifi Password:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().WifiPassword)
	} else {
		fmt.Printf("%-25s", "Wifi Password:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "Wifi AP Mode:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().WifiApMode)
	if len(prefs.GetGetRadioResponse().GetPreferences().Region.String()) > 0 {
		fmt.Printf("%-25s", "Region:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().Region.String())
	} else {
		fmt.Printf("%-25s", "Region:")
		fmt.Printf("%s\n", "N/A")
	}
	if len(prefs.GetGetRadioResponse().GetPreferences().ChargeCurrent.String()) > 0 {
		fmt.Printf("%-25s", "Charge Current:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().ChargeCurrent.String())
	} else {
		fmt.Printf("%-25s", "Charge Current:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "Is router:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().IsRouter)

	fmt.Printf("%-25s", "Is Low Power:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().IsLowPower)

	fmt.Printf("%-25s", "Fixed Position:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().FixedPosition)

	fmt.Printf("%-25s", "Serial Disabled:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().SerialDisabled)
	if len(prefs.GetGetRadioResponse().GetPreferences().LocationShare.String()) > 0 {
		fmt.Printf("%-25s", "Location Share:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().LocationShare.String())
	} else {
		fmt.Printf("%-25s", "Location Share:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "GPS:")
	fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().GpsOperation)

	fmt.Printf("%-25s", "GPS Update Interval:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().GpsUpdateInterval)

	fmt.Printf("%-25s", "GPS Attempt Time:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().GpsAttemptTime)

	fmt.Printf("%-25s", "Frequency Offset:")
	fmt.Printf("%f\n", prefs.GetGetRadioResponse().GetPreferences().FrequencyOffset)
	if len(prefs.GetGetRadioResponse().GetPreferences().MqttServer) > 0 {

		fmt.Printf("%-25s", "Mqtt Server:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().MqttServer)
	} else {
		fmt.Printf("%-25s", "Mqtt Server:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "Mqtt Disabled:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().MqttDisabled)

	return nil
}
