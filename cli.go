package main

import (
	b64 "encoding/base64"
	"flag"
	"fmt"
	"os"

	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
	"google.golang.org/protobuf/proto"
)

// Init starts the CLI and determines flags
func Init() {

	var port string
	var message string
	var owner string
	var url string
	var to int

	flag.StringVar(&port, "port", "", "The serial port for the radio (Required)")
	flag.StringVar(&message, "text", "", "Send a text message")
	flag.StringVar(&url, "url", "", "Set the radio URL")
	flag.StringVar(&owner, "setowner", "", "Set the listed owner for the radio")
	flag.IntVar(&to, "to", 0, "Node to receive text")
	recv := flag.Bool("recv", false, "Wait for new messages")
	infoPtr := flag.Bool("info", false, "Display radio information")
	longslowPtr := flag.Bool("longSlow", false, "Set long-range but slow channel")
	shortFast := flag.Bool("shortFast", false, "Set short-range but fast channel")

	flag.Usage = func() {
		flagSet := flag.CommandLine
		fmt.Printf("A command line tool for interacting with meshtastic radios\n")
		fmt.Printf("\n")
		fmt.Printf("USAGE\n")
		fmt.Printf("meshtastic-go -p <port> [COMMAND]\n")
		fmt.Printf("\n")
		fmt.Printf("COMMANDS\n")
		fmt.Printf("\n")
		order := []string{"port", "text", "info", "setowner", "recv", "url", "longSlow", "shortFast"}
		for _, name := range order {
			flag := flagSet.Lookup(name)
			fmt.Printf("-%s\t", flag.Name)
			fmt.Printf("  %s\n", flag.Usage)
		}
	}

	flag.Parse()

	if port == "" {
		flag.Usage()
		os.Exit(1)
	}

	radio := Radio{}
	if len(message) > 0 || *infoPtr || *recv || len(owner) > 0 || len(url) > 0 || *longslowPtr || *shortFast {
		radio.Init(port)
		defer radio.Close()
	}

	if message != "" {
		err := radio.SendTextMessage(message, to)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}

	//TODO: Add error handling
	if *recv {
		getRecievedMessages(radio)
	}

	//TODO: Add error handling
	if *infoPtr {
		getRadioInfo(radio)
	}

	//TODO: Add error handling
	if *longslowPtr {
		radio.SetChannel(0)
	}

	//TODO: Add error handling
	if *shortFast {
		radio.SetChannel(1)
	}

	if url != "" {
		err := radio.SetChannelURL(url)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		} else {
			fmt.Printf("Set URL successful\n")
		}
	}

	if owner != "" {
		err := radio.SetRadioOwner(owner)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		} else {
			fmt.Printf("Set owner successful\n")
		}
	}

}

func getRecievedMessages(r Radio) {

	fmt.Printf("\n")
	fmt.Printf("Recieved Messages\n")
	fmt.Printf("From\t\t")
	fmt.Printf("To\t\t")
	fmt.Printf("Port Num\t\t")
	fmt.Printf("Payload\t\n")
	for {

		responses, err := r.GetRadioInfo()
		if err != nil {
			fmt.Println(err)
		}

		recievedMessages := make([]*pb.FromRadio_Packet, 0)

		for _, response := range responses {
			if packet, ok := response.GetPayloadVariant().(*pb.FromRadio_Packet); ok {
				recievedMessages = append(recievedMessages, packet)
			}
		}

		if len(recievedMessages) > 0 {
			printMessages(recievedMessages)
		}
	}

}

func getRadioInfo(r Radio) {

	responses, err := r.GetRadioInfo()
	if err != nil {
		fmt.Println(err)
	}

	nodes := make([]*pb.FromRadio_NodeInfo, 0)
	recievedMessages := make([]*pb.FromRadio_Packet, 0)

	for _, response := range responses {

		if info, ok := response.GetPayloadVariant().(*pb.FromRadio_MyInfo); ok {
			fmt.Printf("%-25s", "Node Number: ")
			fmt.Printf("%d\n", info.MyInfo.MyNodeNum)
			fmt.Printf("%-25s", "GPS: ")
			fmt.Printf("%t\n", info.MyInfo.HasGps)
			fmt.Printf("%-25s", "Number of Channels: ")
			fmt.Printf("%d\n", info.MyInfo.NumChannels)
			fmt.Printf("%-25s", "Region: ")
			fmt.Printf("%s\n", info.MyInfo.Region)
			fmt.Printf("%-25s", "Hardware Model: ")
			fmt.Printf("%s\n", info.MyInfo.HwModel)
			fmt.Printf("%-25s", "Firmware: ")
			fmt.Printf("%s\n", info.MyInfo.FirmwareVersion)
			fmt.Printf("%-25s", "Packet ID Bits: ")
			fmt.Printf("%d\n", info.MyInfo.PacketIdBits)
			fmt.Printf("%-25s", "Current Packet ID: ")
			fmt.Printf("%d\n", info.MyInfo.CurrentPacketId)
			fmt.Printf("%-25s", "Node Number of Bits: ")
			fmt.Printf("%d\n", info.MyInfo.NodeNumBits)
			fmt.Printf("%-25s", "Message Timeout (msec): ")
			fmt.Printf("%d\n", info.MyInfo.MessageTimeoutMsec)
			fmt.Printf("%-25s", "Min App Version: ")
			fmt.Printf("%d\n", info.MyInfo.MinAppVersion)

		}

		if radioInfo, ok := response.GetPayloadVariant().(*pb.FromRadio_Radio); ok {
			if radioInfo.Radio.Preferences != nil {
				fmt.Printf("\n")
				fmt.Printf("Preferences=====\n")
				fmt.Printf("%-25s", "ls secs: ")
				fmt.Printf("%d\n", radioInfo.Radio.Preferences.LsSecs)
				fmt.Printf("%-25s", "Region: ")
				fmt.Printf("%d\n", radioInfo.Radio.Preferences.Region)
			}

			if radioInfo.Radio.ChannelSettings != nil {
				fmt.Printf("\n")
				fmt.Printf("Channel Settings:\n")
				fmt.Printf("%-25s", "Modem Config: ")
				fmt.Printf("%s\n", radioInfo.Radio.ChannelSettings.ModemConfig)
				fmt.Printf("%-25s", "PSK: ")
				fmt.Printf("%q\n", radioInfo.Radio.ChannelSettings.Psk)

				protoChannel := radioInfo.Radio.ChannelSettings

				out, err := proto.Marshal(protoChannel)
				if err != nil {
					fmt.Printf("ERROR: Error parsing channel URL")
				}

				url := b64.StdEncoding.EncodeToString(out)

				fmt.Printf("%-25s", "Channel URL: ")
				fmt.Printf("https://www.meshtastic.org/c/#%s\n", url)
			}
		}

		if nodeInfo, ok := response.GetPayloadVariant().(*pb.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}

		if packet, ok := response.GetPayloadVariant().(*pb.FromRadio_Packet); ok {
			recievedMessages = append(recievedMessages, packet)
		}

	}

	if len(nodes) > 0 {
		fmt.Printf("\n")
		fmt.Printf("Nodes in Mesh:\n")

		fmt.Printf("%-80s", "========================================================================================================\n")
		fmt.Printf("| %-20s| ", "Node Number")
		fmt.Printf("%-20s| ", "User")
		fmt.Printf("%-20s| ", "Battery")
		fmt.Printf("%-20s| ", "Latitude")
		fmt.Printf("%s", "Longitude    |\n")
		fmt.Printf("%-80s", "--------------------------------------------------------------------------------------------------------\n")
		for _, node := range nodes {
			if node.NodeInfo != nil {
				fmt.Printf("| %-20s| ", fmt.Sprint(node.NodeInfo.Num))
				fmt.Printf("%-20s| ", node.NodeInfo.User.LongName)
				fmt.Printf("%-20s| ", fmt.Sprint(node.NodeInfo.Position.BatteryLevel))
				fmt.Printf("%-20s| ", fmt.Sprint(node.NodeInfo.Position.LatitudeI))
				fmt.Printf("%s   |\n", fmt.Sprint(node.NodeInfo.Position.LongitudeI))
			}
		}
		fmt.Printf("%-80s", "========================================================================================================\n")
	}

	if len(recievedMessages) > 0 {
		fmt.Printf("\n")
		fmt.Printf("Recieved Messages:\n")
		fmt.Printf("%-80s", "========================================================================================================\n")
		fmt.Printf("| %-20s| ", "From")
		fmt.Printf("%-20s| ", "To")
		fmt.Printf("%-20s| ", "Port Num")
		fmt.Printf("%-20s| ", "Payload")
		printMessages(recievedMessages)
		fmt.Printf("%-80s", "--------------------------------------------------------------------------------------------------------\n")
	}

}

// TODO: Verify formatting
func printMessages(messages []*pb.FromRadio_Packet) {

	for _, message := range messages {
		fmt.Printf("| %-20s| ", fmt.Sprint(message.Packet.From))
		fmt.Printf("%-20s| ", fmt.Sprint(message.Packet.To))
		fmt.Printf("%-20s| ", message.Packet.GetDecoded().GetData().GetPortnum())
		fmt.Printf("%s   |\n", message.Packet.GetDecoded().GetData().Payload)
	}
}

/****************
my_info:{my_node_num:862621917  has_gps:true  num_channels:13  region:"1.0-US"  hw_model:"tbeam"  firmware_version:"1.1.50"  packet_id_bits:32  current_packet_id:1055931951  node_num_bits:32  message_timeout_msec:300000  min_app_version:20120}

radio:{preferences:{ls_secs:300  region:US}  channel_settings:{modem_config:Bw125Cr48Sf4096  psk:"\x01"}}

node_info:{num:862621917  user:{id:"!336a90dd"  long_name:"Unknown 90dd"  short_name:"?DD"  macaddr:"\xc4O3j\x90\xdd"}  position:{battery_level:100  time:1618855568}}

config_complete_id:42

packet:{from:862621917  to:4294967295  decoded:{data:{portnum:NODEINFO_APP  payload:"\n\t!336a90dd\x12\x0cUnknown 90dd\x1a\x03?DD\"\x06\xc4O3j\x90\xdd"}  want_response:true}  id:1055931946  rx_time:1618855478  hop_limit:3  priority:BACKGROUND}
packet:{from:862621917  to:862621917  decoded:{data:{portnum:NODEINFO_APP  payload:"\n\t!336a90dd\x12\x0cUnknown 90dd\x1a\x03?DD\"\x06\xc4O3j\x90\xdd"}}  id:1055931947  rx_time:1618855478  hop_limit:3}
packet:{from:862621917  to:862621917  decoded:{data:{portnum:NODEINFO_APP  payload:"\n\t!336a90dd\x12\x0cUnknown 90dd\x1a\x03?DD\"\x06\xc4O3j\x90\xdd"}}  id:1055931948  rx_time:1618855478  hop_limit:3}
packet:{from:862621917  to:4294967295  decoded:{data:{portnum:POSITION_APP  payload:" dMT\xc6}`"}  want_response:true}  id:1055931949  rx_time:1618855508  hop_limit:3  priority:BACKGROUND}
packet:{from:862621917  to:862621917  decoded:{data:{portnum:POSITION_APP  payload:" dMT\xc6}`"}}  id:1055931950  rx_time:1618855508  hop_limit:3}
packet:{from:862621917  to:862621917  decoded:{data:{portnum:POSITION_APP  payload:" dMT\xc6}`"}}  id:1055931951  rx_time:1618855508  hop_limit:3}
*******/
