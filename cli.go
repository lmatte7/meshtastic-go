package main

import (
	"flag"
	"fmt"
	"os"

	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
)

// Init starts the CLI and determines flags
func Init() {

	var port string
	var message string
	var to int

	flag.StringVar(&port, "port", "", "The serial port for the radio (Required)")
	flag.StringVar(&message, "text", "", "Send a text message")
	flag.IntVar(&to, "to", 0, "Node to receive text")
	recv := flag.Bool("recv", false, "Wait for new messages")
	infoPtr := flag.Bool("info", false, "Display radio information")

	flag.Usage = func() {
		flagSet := flag.CommandLine
		fmt.Printf("A command line tool for interacting with meshtastic radios\n")
		fmt.Printf("\n")
		fmt.Printf("USAGE\n")
		fmt.Printf("meshtastic-go -p <port> [COMMAND]\n")
		fmt.Printf("\n")
		fmt.Printf("COMMANDS\n")
		fmt.Printf("\n")
		order := []string{"port", "text", "info", "recv"}
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
	if len(message) > 0 || *infoPtr || *recv {
		radio.Init(port)
		defer radio.Close()
	}

	if message != "" {
		radio.SendTextMessage(message, to)
	}

	if *recv {
		getRecievedMessages(radio)
	}

	if *infoPtr {
		getRadioInfo(radio)
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
			fmt.Printf("Node Number: %d\n", info.MyInfo.MyNodeNum)
			fmt.Printf("GPS: %t\n", info.MyInfo.HasGps)
			fmt.Printf("Number of Channels: %d\n", info.MyInfo.NumChannels)
			fmt.Printf("Region: %s\n", info.MyInfo.Region)
			fmt.Printf("Hardware Model: %s\n", info.MyInfo.HwModel)
			fmt.Printf("Firmware: %s\n", info.MyInfo.FirmwareVersion)
			fmt.Printf("Packet ID Bits: %d\n", info.MyInfo.PacketIdBits)
			fmt.Printf("Current Packet ID: %d\n", info.MyInfo.CurrentPacketId)
			fmt.Printf("Node Number of Bits: %d\n", info.MyInfo.NodeNumBits)
			fmt.Printf("Message Timeout (msec): %d\n", info.MyInfo.MessageTimeoutMsec)
			fmt.Printf("Min App Version: %d\n", info.MyInfo.MinAppVersion)

		}

		if radioInfo, ok := response.GetPayloadVariant().(*pb.FromRadio_Radio); ok {
			fmt.Printf("\n")
			fmt.Printf("Preferences:\n")
			fmt.Printf("ls secs: %d\n", radioInfo.Radio.Preferences.LsSecs)
			fmt.Printf("Region: %d\n", radioInfo.Radio.Preferences.Region)
		}

		if channelInfo, ok := response.GetPayloadVariant().(*pb.FromRadio_Channel); ok {
			fmt.Printf("\n")
			fmt.Printf("Channel Settings:\n")
			fmt.Printf("Modem Config: %s\n", channelInfo.Channel.ModemConfig)
			fmt.Printf("PSK: %s\n", channelInfo.Channel.Psk)
		}

		if nodeInfo, ok := response.GetPayloadVariant().(*pb.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}

		if packet, ok := response.GetPayloadVariant().(*pb.FromRadio_Packet); ok {
			recievedMessages = append(recievedMessages, packet)
		}

	}

	// TODO: Catch segmentation violation for unitinitalized radios
	if len(nodes) > 0 {
		fmt.Printf("\n")
		fmt.Printf("Nodes in Mesh:\n")

		fmt.Printf("Num\t\t")
		fmt.Printf("User\t\t")
		fmt.Printf("Battery\t\t")
		fmt.Printf("Latitude\t")
		fmt.Printf("Longitude\t\n")
		for _, node := range nodes {
			fmt.Printf("%d\t", node.NodeInfo.Num)
			fmt.Printf("%s\t", node.NodeInfo.User.Id)
			fmt.Printf("%d\t\t", node.NodeInfo.Position.BatteryLevel)
			fmt.Printf("%d\t\t", node.NodeInfo.Position.LatitudeI)
			fmt.Printf("%d\t\n", node.NodeInfo.Position.LongitudeI)
		}
	}

	if len(recievedMessages) > 0 {
		fmt.Printf("\n")
		fmt.Printf("Recieved Messages\n")
		fmt.Printf("From\t\t")
		fmt.Printf("To\t\t")
		fmt.Printf("Port Num\t\t")
		fmt.Printf("Payload\t\n")
		printMessages(recievedMessages)
	}

}

func printMessages(messages []*pb.FromRadio_Packet) {

	for _, message := range messages {
		fmt.Printf("%d\t", message.Packet.From)
		fmt.Printf("%d\t", message.Packet.To)
		fmt.Printf("%s\t", message.Packet.GetDecoded().GetData().GetPortnum())
		fmt.Printf("%s\t\n", message.Packet.GetDecoded().GetData().Payload)
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
