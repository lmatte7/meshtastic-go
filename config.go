package main

import (
	"fmt"
	"reflect"

	"github.com/lmatte7/gomesh"
	"github.com/urfave/cli/v2"
)

// setConfig sets a radio configuration value on the radio
func setConfig(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	err := radio.SetRadioConfig(c.String("key"), c.String("value"))
	if err != nil {
		return cli.Exit(err, 0)
	}

	fmt.Printf("%s set successfully\n", c.String("key"))
	return nil
}

func setOwner(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	return radio.SetRadioOwner(c.String("name"))
}

func printConfig(r gomesh.Radio) error {

	configSettings, moduleSettings, err := r.GetRadioConfig()
	if err != nil {
		return err
	}

	for _, config := range configSettings {
		if deviceConfig := config.Config.GetDevice(); deviceConfig != nil {
			fmt.Printf("deviceConfig: %+v\n", *deviceConfig)
			continue
		}

		if deviceConfig := config.Config.GetPosition(); deviceConfig != nil {
			fmt.Printf("positionConfig: %+v\n", *deviceConfig)
			continue
		}

		if deviceConfig := config.Config.GetPower(); deviceConfig != nil {
			fmt.Printf("powerConfig: %+v\n", *deviceConfig)
			continue
		}

		if deviceConfig := config.Config.GetNetwork(); deviceConfig != nil {
			fmt.Printf("networkConfig: %+v\n", *deviceConfig)
			continue
		}

		if deviceConfig := config.Config.GetDisplay(); deviceConfig != nil {
			fmt.Printf("displayConfig: %+v\n", *deviceConfig)
			continue
		}

		if deviceConfig := config.Config.GetLora(); deviceConfig != nil {
			fmt.Printf("loraConfig: %+v\n", *deviceConfig)
			continue
		}

		if deviceConfig := config.Config.GetBluetooth(); deviceConfig != nil {
			fmt.Printf("bluetoothConfig: %+v\n", *deviceConfig)
			continue
		}

	}

	return nil

	fmt.Printf("Radio Config:\n")
	fmt.Printf(
		"%-40s",
		"==============================================================================\n",
	)
	for _, config := range configSettings {
		if deviceConfig := config.Config.GetDevice(); deviceConfig != nil {
			printSection("Device Config Options", *deviceConfig)
		} else if deviceConfig := config.Config.GetPosition(); deviceConfig != nil {
			printSection("Position Config Options", *deviceConfig)
		} else if deviceConfig := config.Config.GetPower(); deviceConfig != nil {
			printSection("Power Config Options", *deviceConfig)
		} else if deviceConfig := config.Config.GetNetwork(); deviceConfig != nil {
			printSection("Network Config Options", *deviceConfig)
		} else if deviceConfig := config.Config.GetDisplay(); deviceConfig != nil {
			printSection("Display Config Options", *deviceConfig)
		} else if deviceConfig := config.Config.GetLora(); deviceConfig != nil {
			printSection("Lora Config Options", *deviceConfig)
		} else if deviceConfig := config.Config.GetBluetooth(); deviceConfig != nil {
			printSection("Bluetooth Config Options", *deviceConfig)
		}
	}

	for _, module := range moduleSettings {

		if moduleConfig := module.ModuleConfig.GetMqtt(); moduleConfig != nil {
			printSection("Mqtt Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetSerial(); moduleConfig != nil {
			printSection("Serial Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetExternalNotification(); moduleConfig != nil {
			printSection(
				"External Notification Module Options",
				*moduleConfig,
			)
		}
		if moduleConfig := module.ModuleConfig.GetStoreForward(); moduleConfig != nil {
			printSection("Store Forward Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetRangeTest(); moduleConfig != nil {
			printSection("Range Test Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetTelemetry(); moduleConfig != nil {
			printSection("Telemetry Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetCannedMessage(); moduleConfig != nil {
			printSection("Canned Message Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetAudio(); moduleConfig != nil {
			printSection("Audio Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetRemoteHardware(); moduleConfig != nil {
			printSection("Serial Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetNeighborInfo(); moduleConfig != nil {
			printSection("Neighbor Info Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetAmbientLighting(); moduleConfig != nil {
			printSection("Ambient Lighting Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetDetectionSensor(); moduleConfig != nil {
			printSection("Detection Sensor Module Options", *moduleConfig)
		}
		if moduleConfig := module.ModuleConfig.GetPaxcounter(); moduleConfig != nil {
			printSection("Pax Counter Module Options", *moduleConfig)
		}
	}

	return nil
}

func showRadioConfig(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	return printConfig(radio)
}

func printSection(title string, t interface{}) {
	v := reflect.ValueOf(t)
	fmt.Printf(
		"\n%-40s",
		"-------------------------------------------------\n",
	)
	fmt.Printf("%-48s|\n", title)
	fmt.Printf(
		"%-40s",
		"-------------------------------------------------\n",
	)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			fmt.Printf("%-40s", v.Type().Field(i).Name)
			fmt.Printf("%v\n", v.Field(i))
		}
	}
}
