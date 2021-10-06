package main

import (
	"fmt"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
)

func showRadioInfo(c *cli.Context) error {

	radio := getRadio(c)
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

func showAllRadioInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	err := getRadioInfo(radio)
	if err != nil {
		return err
	}

	return nil

}

func showNodeInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	err := displayNodes(radio)
	if err != nil {
		return err
	}

	return nil

}

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

	fmt.Printf("%-80s", "=======================================================================================================\n")
	fmt.Printf("| %-15s| ", "Node Number")
	fmt.Printf("%-15s| ", "User")
	fmt.Printf("%-15s| ", "Battery")
	fmt.Printf("%-15s| ", "Altitude")
	fmt.Printf("%-15s| ", "Latitude")
	fmt.Printf("%-15s", "Longitude      |\n")
	fmt.Printf("%-80s", "-------------------------------------------------------------------------------------------------------\n")
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
				fmt.Printf("%-15s| ", fmt.Sprint(node.NodeInfo.Position.Altitude))
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
	fmt.Printf("%-80s", "=======================================================================================================\n")
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
