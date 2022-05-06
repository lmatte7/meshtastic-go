package main

import (
	"fmt"
	"reflect"

	"github.com/lmatte7/gomesh"
	"github.com/urfave/cli/v2"
)

func setPref(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	return radio.SetUserPreferences(c.String("key"), c.String("value"))
}

func setOwner(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	return radio.SetRadioOwner(c.String("name"))
}

func printPreferences(r gomesh.Radio) error {

	prefs, _ := r.GetRadioPreferences()
	prefs, err := r.GetRadioPreferences()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(*prefs.GetGetRadioResponse().GetPreferences())
	fmt.Printf("%s", "\n")
	fmt.Printf("Radio Preferences:\n")
	fmt.Printf("%-40s", "==============================================================================\n")
	fmt.Printf("%-55s| ", "Field Name")
	fmt.Printf("%-20s|\n", "Current Value")
	fmt.Printf("%-40s", "------------------------------------------------------------------------------\n")

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			fmt.Printf("%-55s| ", v.Type().Field(i).Name)
			fmt.Printf("%-20v|\n", v.Field(i))
			// fmt.Printf("%s - %s - %v\n", v.Type().Field(i).Name, v.Type().Field(i).Type.Kind(), v.Field(i))
		}
	}
	fmt.Printf("%-40s", "==============================================================================\n")

	return nil
}

func printRadioPreferences(r gomesh.Radio) error {

	prefs, _ := r.GetRadioPreferences()
	prefs, err := r.GetRadioPreferences()
	if err != nil {
		return err
	}

	fmt.Printf("%-25s", "SendOwnerInterval:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().SendOwnerInterval)
	fmt.Printf("%-25s", "WaitBluetoothSecs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().WaitBluetoothSecs)
	fmt.Printf("%-25s", "ScreenOnSecs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().ScreenOnSecs)
	fmt.Printf("%-25s", "PhoneTimeoutSecs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().PhoneTimeoutSecs)
	fmt.Printf("%-25s", "PhoneSdsTimeoutSec:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().PhoneSdsTimeoutSec)
	fmt.Printf("%-25s", "MeshSdsTimeoutSecs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().MeshSdsTimeoutSecs)
	fmt.Printf("%-25s", "SdsSecs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().SdsSecs)
	fmt.Printf("%-25s", "LsSecs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().LsSecs)
	fmt.Printf("%-25s", "MinWakeSecs:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().MinWakeSecs)

	if len(prefs.GetGetRadioResponse().GetPreferences().WifiSsid) > 0 {
		fmt.Printf("%-25s", "WifiSsid:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().WifiSsid)
	} else {
		fmt.Printf("%-25s", "WifiSsid:")
		fmt.Printf("%s\n", "N/A")
	}
	if len(prefs.GetGetRadioResponse().GetPreferences().WifiPassword) > 0 {
		fmt.Printf("%-25s", "WifiPassword:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().WifiPassword)
	} else {
		fmt.Printf("%-25s", "WifiPassword:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "WifiApMode:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().WifiApMode)
	if len(prefs.GetGetRadioResponse().GetPreferences().Region.String()) > 0 {
		fmt.Printf("%-25s", "Region:")
		fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().Region)
	} else {
		fmt.Printf("%-25s", "Region:")
		fmt.Printf("%s\n", "N/A")
	}
	if len(prefs.GetGetRadioResponse().GetPreferences().ChargeCurrent.String()) > 0 {
		fmt.Printf("%-25s", "ChargeCurrent:")
		fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().ChargeCurrent)
	} else {
		fmt.Printf("%-25s", "ChargeCurrent:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "IsRouter:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().IsRouter)

	fmt.Printf("%-25s", "IsLowPower:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().IsLowPower)

	fmt.Printf("%-25s", "FixedPosition:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().FixedPosition)

	fmt.Printf("%-25s", "SerialDisabled:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().SerialDisabled)
	if len(prefs.GetGetRadioResponse().GetPreferences().LocationShare.String()) > 0 {
		fmt.Printf("%-25s", "LocationShare:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().LocationShare.String())
	} else {
		fmt.Printf("%-25s", "LocationShare:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "GpsAccept_2D:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().GpsAccept_2D)

	fmt.Printf("%-25s", "IsAlwaysPowered:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().IsAlwaysPowered)

	fmt.Printf("%-25s", "GpsMaxDop:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().GpsMaxDop)

	fmt.Printf("%-25s", "IsRouter:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().IsRouter)

	fmt.Printf("%-25s", "GpsUpdateInterval:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().GpsUpdateInterval)

	fmt.Printf("%-25s", "GpsAttemptTime:")
	fmt.Printf("%d\n", prefs.GetGetRadioResponse().GetPreferences().GpsAttemptTime)

	fmt.Printf("%-25s", "FrequencyOffset:")
	fmt.Printf("%f\n", prefs.GetGetRadioResponse().GetPreferences().FrequencyOffset)
	if len(prefs.GetGetRadioResponse().GetPreferences().MqttServer) > 0 {

		fmt.Printf("%-25s", "MqttServer:")
		fmt.Printf("%s\n", prefs.GetGetRadioResponse().GetPreferences().MqttServer)
	} else {
		fmt.Printf("%-25s", "MqttServer:")
		fmt.Printf("%s\n", "N/A")
	}

	fmt.Printf("%-25s", "MqttDisabled:")
	fmt.Printf("%t\n", prefs.GetGetRadioResponse().GetPreferences().MqttDisabled)

	return nil
}

func showRadioPreferences(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	return printPreferences(radio)
}
