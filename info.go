package main

import (
	"fmt"
	"reflect"

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

	printRadioInfo(responses)

	return nil

}

func showAllRadioInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return getRadioInfo(radio)
}

func showNodeInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return displayNodes(radio)
}

func showPositionInfo(c *cli.Context) error {

	positionPacket := &gomeshproto.FromRadio{}

	r := getRadio(c)
	defer r.Close()

	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	for _, packet := range responses {
		if config := packet.GetConfig(); config != nil {
			if gpsConfig := config.GetPosition(); gpsConfig != nil {
				positionPacket = packet
			}
		}
	}

	displayPositionInfo(positionPacket)

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

	printRadioInfo(responses)

	return nil
}

func printNodes(nodes []*gomeshproto.FromRadio_NodeInfo) {
	fmt.Printf("\n")
	fmt.Printf("Nodes in Mesh:\n")

	printDoubleDivider()
	fmt.Printf("| %-15s| ", "Node Number")
	fmt.Printf("%-15s| ", "User")
	fmt.Printf("%-15s| ", "Battery")
	fmt.Printf("%-15s| ", "Altitude")
	fmt.Printf("%-15s| ", "Latitude")
	fmt.Printf("%-15s", "Longitude      |\n")
	printSingleDivider()
	for _, node := range nodes {
		if node.NodeInfo != nil {
			fmt.Printf("| %-15s| ", fmt.Sprint(node.NodeInfo.Num))
			if node.NodeInfo.User != nil {
				fmt.Printf("%-15s| ", node.NodeInfo.User.LongName)
			} else {
				fmt.Printf("%-15s| ", "N/A")
			}
			if node.NodeInfo.Position != nil {
				fmt.Printf("%-15s| ", fmt.Sprint(node.NodeInfo.DeviceMetrics.BatteryLevel))
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
	printDoubleDivider()
}

func printRadioInfo(info []*gomeshproto.FromRadio) {
	fmt.Printf("%s", "\nRadio Settings: \n")
	nodes := make([]*gomeshproto.FromRadio_NodeInfo, 0)
	channels := make([]*gomeshproto.FromRadio_Channel, 0)
	positionPacket := &gomeshproto.FromRadio{}

	for _, packet := range info {
		if nodeInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}
		if channelInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_Channel); ok {
			channels = append(channels, channelInfo)
		}
		if config := packet.GetConfig(); config != nil {
			if gpsConfig := config.GetPosition(); gpsConfig != nil {
				positionPacket = packet
			}
		}
		if metaInfo := packet.GetMetadata(); metaInfo != nil {
			fmt.Printf("%s", "Radio Metadata\n")
			v := reflect.ValueOf(*metaInfo)
			for i := 0; i < v.NumField(); i++ {
				if v.Field(i).CanInterface() {
					fmt.Printf("%-25s", v.Type().Field(i).Name)
					fmt.Printf("%v\n", v.Field(i))
				}
			}
		}
		if nodeInfo := packet.GetNodeInfo(); nodeInfo != nil {
			fmt.Printf("%s", "\n\nNode Info\n")
			v := reflect.ValueOf(*nodeInfo)
			for i := 0; i < v.NumField(); i++ {
				fmt.Printf("%-25s", v.Type().Field(i).Name)
				fmt.Printf("%v\n", v.Field(i))
			}
		}
	}

	displayPositionInfo(positionPacket)
	printNodes(nodes)
	printChannels(channels)

}

func displayPositionInfo(packet *gomeshproto.FromRadio) {
	if config := packet.GetConfig(); config != nil {
		if gpsConfig := config.GetPosition(); gpsConfig != nil {
			fmt.Printf("%s", "\n\nPosition Settings\n")
			v := reflect.ValueOf(*gpsConfig)
			for i := 0; i < v.NumField(); i++ {
				if v.Field(i).CanInterface() {
					fmt.Printf("%-35s", v.Type().Field(i).Name)
					fmt.Printf("%v\n", v.Field(i))
				}
			}
		}
	}
}
