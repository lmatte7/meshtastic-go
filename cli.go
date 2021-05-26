package main

import (
	b64 "encoding/base64"
	"flag"
	"fmt"
	"os"
	"regexp"

	pb "github.com/lmatte7/meshtastic-go/go-meshtastic-protobufs"
	"google.golang.org/protobuf/proto"
)

// Init starts the CLI and determines flags
func Init() {

	var port string
	var message string
	var owner string
	var setKey string
	var setValue string
	var url string
	var to int64

	flag.StringVar(&port, "port", "", "--port=port The serial port for the radio (Required)")
	flag.StringVar(&message, "text", "", "--text=message Send a text message")
	flag.StringVar(&url, "url", "", "--url=channel_curl Set the radio URL")
	flag.StringVar(&setKey, "setKey", "", "--setKey=key The key to set for a custom user preference option. Used with setValue")
	flag.StringVar(&setValue, "setValue", "", "--setValue=value The value to set for a custom user preference option. Used with setKey")
	flag.StringVar(&owner, "setOwner", "", "--setowner=owner Set the listed owner for the radio")
	flag.Int64Var(&to, "to", 0, "--to=destination Node to receive text (Used with sendtext)")
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
		order := []string{"port", "info", "recv", "longSlow", "shortFast", "text", "setOwner", "setKey", "setValue"}
		for _, name := range order {
			flag := flagSet.Lookup(name)
			fmt.Printf("--%-15s", flag.Name)
			fmt.Printf("%-10s\n", flag.Usage)
		}
	}

	flag.Parse()

	if port == "" {
		flag.Usage()
		os.Exit(1)
	}

	radio := Radio{}
	if len(message) > 0 || *infoPtr || *recv || len(owner) > 0 || len(url) > 0 || *longslowPtr || *shortFast || len(setKey) > 0 || len(setValue) > 0 {
		radio.Init(port)
		defer radio.Close()
	}

	if message != "" {
		err := radio.SendTextMessage(message, to)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}

	if *recv {
		err := getRecievedMessages(radio)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}

	if *infoPtr {
		err := getRadioInfo(radio)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}

	if *longslowPtr {
		err := radio.SetChannel(0)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		} else {
			fmt.Printf("Set channel\n")
		}
	}

	if *shortFast {
		err := radio.SetChannel(1)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		} else {
			fmt.Printf("Set channel\n")
		}
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

	if setValue != "" && setKey != "" {
		err := radio.SetUserPreferences(setKey, setValue)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		} else {
			fmt.Printf("Set user preferences successful\n")
		}
	}

}

func getRecievedMessages(r Radio) error {

	printMessageHeader()
	for {

		responses, err := r.GetRadioInfo()
		if err != nil {
			return err
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

func getRadioInfo(r Radio) error {

	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
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
				fmt.Printf("Preferences:\n")
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

func printMessages(messages []*pb.FromRadio_Packet) {

	for _, message := range messages {
		fmt.Printf("| %-15s| ", fmt.Sprint(message.Packet.From))
		fmt.Printf("%-15s| ", fmt.Sprint(message.Packet.To))
		fmt.Printf("%-18s| ", message.Packet.GetDecoded().GetData().GetPortnum())
		re := regexp.MustCompile(`\r?\n`)
		escMesg := re.ReplaceAllString(string(message.Packet.GetDecoded().GetData().Payload), "")
		fmt.Printf("%-53s", escMesg)
		fmt.Printf("%s", "|\n")
	}
}
