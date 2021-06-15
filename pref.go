package main

import (
	"fmt"

	"github.com/lmatte7/gomesh"
	"github.com/urfave/cli/v2"
)

func setPref(c *cli.Context) error {
	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = radio.SetUserPreferences(c.String("key"), c.String("value"))
	if err != nil {
		return err
	}

	return nil
}

func setOwner(c *cli.Context) error {
	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = radio.SetRadioOwner(c.String("owner"))
	if err != nil {
		return err
	}

	return nil
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

func showRadioPreferences(c *cli.Context) error {
	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	err = printRadioPreferences(radio)
	if err != nil {
		return err
	}

	return nil
}
