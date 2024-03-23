package main

import (
	"github.com/urfave/cli/v2"
)

// TODO: Note  - Since all of the responses are sent at once, can we change to one general set command for channels and preferences???

// TODO: Redo preferences to match new modules
// func setPref(c *cli.Context) error {
// radio := getRadio(c)
// defer radio.Close()

// prefs, _ := radio.GetRadioPreferences()
// prefs, err := radio.GetRadioPreferences()
// if err != nil {
// 	return err
// }

// // v := reflect.ValueOf(*prefs.GetGetRadioResponse().GetPreferences())
// v := reflect.ValueOf(*prefs.GetGetRadioResponse().GetPreferences())

// fieldFound := false
// for i := 0; i < v.NumField(); i++ {
// 	if v.Field(i).CanInterface() {
// 		if v.Type().Field(i).Name == c.String("key") {
// 			fieldFound = true
// 		}
// 	}
// }

// if !fieldFound {
// 	return cli.Exit("Invalid key provided", 0)
// }

// err = radio.SetUserPreferences(c.String("key"), c.String("value"))

// if err != nil {
// 	return cli.Exit(err.Error(), 0)
// }

// fmt.Printf("%s set successfully\n", c.String("key"))
// return nil
// }

func setOwner(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	return radio.SetRadioOwner(c.String("name"))
}

// func printPreferences(r gomesh.Radio) error {

// 	prefs, _ := r.GetRadioPreferences()
// 	prefs, err := r.GetRadioPreferences()
// 	if err != nil {
// 		return err
// 	}

// 	v := reflect.ValueOf(*prefs.GetGetRadioResponse().GetPreferences())
// 	fmt.Printf("%s", "\n")
// 	fmt.Printf("Radio Preferences:\n")
// 	fmt.Printf("%-40s", "==============================================================================\n")
// 	fmt.Printf("%-55s| ", "Field Name")
// 	fmt.Printf("%-20s|\n", "Current Value")
// 	fmt.Printf("%-40s", "------------------------------------------------------------------------------\n")

// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanInterface() {
// 			fmt.Printf("%-55s| ", v.Type().Field(i).Name)
// 			fmt.Printf("%-20v|\n", v.Field(i))
// 			// fmt.Printf("%s - %s - %v\n", v.Type().Field(i).Name, v.Type().Field(i).Type.Kind(), v.Field(i))
// 		}
// 	}
// 	fmt.Printf("%-40s", "==============================================================================\n")

// 	return nil
// }

// func showRadioPreferences(c *cli.Context) error {
// 	radio := getRadio(c)
// 	defer radio.Close()

// return printPreferences(radio)
// }
