package main

import (
	"fmt"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
)

func showMetricInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return displayMetrics(radio)
}

func displayMetrics(r gomesh.Radio) error {
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

	printMetrics(nodes)

	return nil
}

func printMetrics(nodes []*gomeshproto.FromRadio_NodeInfo) {
	fmt.Printf("\n")
	fmt.Printf("Mesh Device Metrics:\n")

	printDoubleDivider()
	fmt.Printf("| %-15s| ", "Node Number")
	fmt.Printf("%-15s| ", "Battery")
	fmt.Printf("%-15s| ", "Voltage")
	fmt.Printf("%-20s| ", "Channel Utilization")
	fmt.Printf("%-15s", "AirUtilTx      |\n")
	printSingleDivider()
	for _, node := range nodes {
		if node.NodeInfo.DeviceMetrics != nil {
			fmt.Printf("| %-15s| ", fmt.Sprint(node.NodeInfo.Num))
			fmt.Printf("%-15s| ", fmt.Sprint(node.NodeInfo.DeviceMetrics.BatteryLevel))
			fmt.Printf("%-15s| ", fmt.Sprint(node.NodeInfo.DeviceMetrics.Voltage))
			fmt.Printf("%-20s| ", fmt.Sprint(node.NodeInfo.DeviceMetrics.ChannelUtilization))
			fmt.Printf("%-15s| \n", fmt.Sprint(node.NodeInfo.DeviceMetrics.AirUtilTx))
		}
	}
	printDoubleDivider()
}
