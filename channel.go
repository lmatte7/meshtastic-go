package main

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"

	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
)

func showChannelInfo(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	info, err := radio.GetRadioInfo()
	if err != nil {
		return cli.Exit("Failed parsing channels", 0)
	}

	channels := make([]*gomeshproto.FromRadio_Channel, 0)

	for _, packet := range info {
		if channelInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_Channel); ok {
			channels = append(channels, channelInfo)
		}
	}

	err = printChannels(channels)
	if err != nil {
		return cli.Exit(err, 0)
	}

	return nil
}

func printChannels(channels []*gomeshproto.FromRadio_Channel) error {

	primaryChannelSettings := gomeshproto.ChannelSettings{}
	allChannelSettings := []*gomeshproto.ChannelSettings{}
	channelSet := gomeshproto.ChannelSet{}

	fmt.Printf("%s", "\n")
	fmt.Printf("Channel Settings:\n")
	printDoubleDivider()
	fmt.Printf("| %-15s| ", "Name")
	fmt.Printf("%-15s| ", "Index")
	fmt.Printf("%-10s| ", "Uplink")
	fmt.Printf("%-10s| ", "Downlink")
	fmt.Printf("%-15s| ", "Role")
	fmt.Printf("%-15s| ", "Precision")
	fmt.Printf("%-90s", "PSK")
	fmt.Printf("%s", "|\n")
	printSingleDivider()
	for _, channelInfo := range channels {

		// fmt.Printf("Channel Info: %v\n\n", channelInfo.Channel.Settings)
		if channelInfo.Channel.GetRole() == gomeshproto.Channel_DISABLED {
			break
		}

		if channelInfo.Channel.GetRole() == gomeshproto.Channel_PRIMARY {
			primaryChannelSettings = *channelInfo.Channel.Settings
		}

		allChannelSettings = append(allChannelSettings, channelInfo.Channel.Settings)

		if len(channelInfo.Channel.Settings.Name) > 0 {
			fmt.Printf("| %-15s| ", channelInfo.Channel.Settings.Name)
		} else {
			fmt.Printf("| %-15s| ", "Default")
		}
		if channelInfo.Channel.Index > 0 {
			fmt.Printf("%-15d| ", channelInfo.Channel.Index)
		} else if channelInfo.Channel.GetRole() == gomeshproto.Channel_PRIMARY {
			fmt.Printf("%-15s| ", "0")
		} else {
			fmt.Printf("%-15s| ", "N/A")
		}
		if channelInfo.Channel.Settings.UplinkEnabled {
			fmt.Printf("%-10s| ", "True")
		} else {
			fmt.Printf("%-10s| ", "False")
		}

		if channelInfo.Channel.Settings.DownlinkEnabled {
			fmt.Printf("%-10s| ", "True")
		} else {
			fmt.Printf("%-10s| ", "False")
		}

		if len(channelInfo.Channel.Role.String()) > 0 {
			fmt.Printf("%-15s| ", channelInfo.Channel.Role.String())
		} else {
			fmt.Printf("%-15s| ", "N/A")
		}
		if channelInfo.Channel.Settings.ModuleSettings.GetPositionPrecision() > 0 {
			fmt.Printf("%-15d| ", channelInfo.Channel.Settings.ModuleSettings.PositionPrecision)
		} else {
			fmt.Printf("%-15s| ", "N/A")
		}
		if len(channelInfo.Channel.Settings.Psk) > 0 {
			re := regexp.MustCompile(`\r?\n`)
			escMesg := re.ReplaceAllString(string(channelInfo.Channel.Settings.Psk), "")
			fmt.Printf("%-90q", escMesg)
		} else {
			fmt.Printf("%-53s| ", "N/A")
		}
		fmt.Printf("%s", "|\n")

	}
	printDoubleDivider()

	channelSet.Settings = allChannelSettings

	out, err := proto.Marshal(&primaryChannelSettings)
	if err != nil {
		return cli.Exit("Error parsing channel URL", 0)
	}

	url := base64.RawURLEncoding.EncodeToString(out)

	fmt.Printf("%-25s", "Primary Channel URL: ")
	fmt.Printf("https://www.meshtastic.org/c/#%s\n", url)

	out, err = proto.Marshal(&channelSet)
	if err != nil {
		return cli.Exit("Error parsing channel URL", 0)
	}

	url = base64.RawURLEncoding.EncodeToString(out)

	fmt.Printf("%-25s", "Full Channel URL: ")
	fmt.Printf("https://www.meshtastic.org/c/#%s\n", url)

	return nil
}

func showChannelOptions(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	info, err := radio.GetRadioInfo()
	if err != nil {
		return cli.Exit("Failed to retrive options", 0)
	}

	for _, packet := range info {
		if channelInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_Channel); ok {
			fmt.Printf("%s", "\nGeneric Channel Options\n")
			printDoubleDivider()
			v := reflect.ValueOf(*channelInfo.Channel)
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					if v.Type().Field(i).Name == "Settings" {
						fmt.Println("\nChannel Setting Options")
						printDoubleDivider()
						cv := reflect.ValueOf(*channelInfo.Channel.Settings)
						for j := 0; j < cv.NumField(); j++ {
							if cv.Type().Field(j).IsExported() {
								if cv.Type().Field(j).Name == "ModuleSettings" {
									fmt.Println("\nModule Setting Options")
									printDoubleDivider()
									mv := reflect.ValueOf(*channelInfo.Channel.Settings.ModuleSettings)
									for k := 0; k < mv.NumField(); k++ {
										if mv.Type().Field(k).IsExported() {
											fmt.Printf("%v\n", mv.Type().Field(k).Name)
										}
									}
								} else {
									fmt.Printf("%v\n", cv.Type().Field(j).Name)
								}
							}
						}
					} else if v.Type().Field(i).Name == "Role" {
						fmt.Println("\nTo set a channel as the primary role set it to index 0")
					} else {
						fmt.Printf("%v\n", v.Type().Field(i).Name)
					}
				}

			}

			break
		}
	}

	return nil
}

func addChannel(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	err := radio.AddChannel(c.String("name"), c.Int("index"))
	if err != nil {
		return cli.Exit(err, 0)
	}

	return nil
}

func deleteChannel(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return radio.DeleteChannel(c.Int("index"))
}

func setChannel(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	err := radio.SetChannel(c.Int("index"), c.String("key"), c.String("value"))

	if err != nil {
		return cli.Exit(err, 0)
	}

	return nil
}

func setUrl(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	err := radio.SetChannelURL(c.String("url"))
	if err != nil {
		return cli.Exit(err, 0)
	}

	return nil
}
