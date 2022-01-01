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
	radio := getRadio(c)
	defer radio.Close()

	return printChannelSettings(radio)
}

func printChannelSettings(r gomesh.Radio) error {
	channels := []gomeshproto.AdminMessage{}
	channelSettings := []*gomeshproto.ChannelSettings{}
	primaryChannelSettings := []*gomeshproto.ChannelSettings{}
	channelSet := gomeshproto.ChannelSet{}
	primaryChannelSet := gomeshproto.ChannelSet{}
	channelCount := 0
	for {

		info, err := r.GetChannelInfo(channelCount)
		if err != nil {
			return err
		}
		if info.GetGetChannelResponse() == nil || info.GetGetChannelResponse().Role == gomeshproto.Channel_DISABLED {
			break
		}

		channels = append(channels, info)
		// Add a guarenteed exit for the loop since there can't be more than 20 channels
		channelCount++
		if channelCount > 20 {
			break
		}
	}

	// Try again if no settings were found
	if len(channels) == 0 {
		for {

			info, err := r.GetChannelInfo(channelCount)
			if err != nil {
				return err
			}
			if info.GetGetChannelResponse() == nil || info.GetGetChannelResponse().Role == gomeshproto.Channel_DISABLED {
				break
			}

			channels = append(channels, info)
			// Add a guarenteed exit for the loop since there can't be more than 20 channels
			channelCount++
			if channelCount > 20 {
				break
			}
		}
	}

	fmt.Printf("%s", "\n")
	fmt.Printf("Channel Settings:\n")
	fmt.Printf("%-80s", "=================================================================================================================================================================\n")
	fmt.Printf("| %-15s| ", "Name")
	fmt.Printf("%-15s| ", "Index")
	fmt.Printf("%-15s| ", "Role")
	fmt.Printf("%-15s| ", "Modem")
	fmt.Printf("%-90s", "PSK")
	fmt.Printf("%s", "|\n")
	fmt.Printf("%-80s", "-----------------------------------------------------------------------------------------------------------------------------------------------------------------\n")
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
		if channelInfo.GetGetChannelResponse().GetIndex() > 0 {
			fmt.Printf("%-15d| ", channelInfo.GetGetChannelResponse().GetIndex())
		} else if channelInfo.GetGetChannelResponse().Role.String() == "PRIMARY" {
			fmt.Printf("%-15s| ", "0")
		} else {
			fmt.Printf("%-15s| ", "N/A")
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
	fmt.Printf("%-80s", "=================================================================================================================================================================\n")

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

func addChannel(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return radio.AddChannel(c.String("name"), c.Int("index"))
}

func deleteChannel(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return radio.DeleteChannel(c.Int("index"))
}

func setChannel(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return radio.SetChannel(c.Int("index"), c.String("key"), c.String("value"))
}

func setUrl(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return radio.SetChannelURL(c.String("url"))
}
