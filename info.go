package main

import (
	"fmt"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
)

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
