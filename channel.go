package main

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
)

func showChannelInfo(c *cli.Context) error {
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

func printChannelSettings(r gomesh.Radio) error {

	channels := make([]gomeshproto.AdminMessage, 0)
	channelSettings := make([]*gomeshproto.ChannelSettings, 0)
	primaryChannelSettings := make([]*gomeshproto.ChannelSettings, 0)
	channelSet := gomeshproto.ChannelSet{}
	primaryChannelSet := gomeshproto.ChannelSet{}
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
		if channelInfo.GetGetChannelResponse().Role == gomeshproto.Channel_PRIMARY {
			primaryChannelSettings = append(primaryChannelSettings, channelInfo.GetGetChannelResponse().GetSettings())
		}

		channelSettings = append(channelSettings, channelInfo.GetGetChannelResponse().GetSettings())

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

	channelSet.Settings = channelSettings
	primaryChannelSet.Settings = primaryChannelSettings

	out, err := proto.Marshal(&primaryChannelSet)
	if err != nil {
		fmt.Printf("ERROR: Error parsing channel URL")
	}

	url := base64.RawURLEncoding.EncodeToString(out)

	fmt.Printf("%-25s", "Primary Channel URL: ")
	fmt.Printf("https://www.meshtastic.org/c/#%s\n", url)

	out, err = proto.Marshal(&channelSet)
	if err != nil {
		fmt.Printf("ERROR: Error parsing channel URL")
	}

	url = base64.RawURLEncoding.EncodeToString(out)

	fmt.Printf("%-25s", "Full Channel URL: ")
	fmt.Printf("https://www.meshtastic.org/c/#%s\n", url)

	return nil
}
