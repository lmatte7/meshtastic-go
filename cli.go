package main

import (
	"flag"
	"fmt"
	"os"
)

// Init starts the CLI and determines flags
func Init() {

	var port string
	var message string

	flag.StringVar(&port, "port", "", "The serial port for the radio (Required)")
	flag.StringVar(&message, "text", "", "Send a text message")
	infoPtr := flag.Bool("info", false, "Display radio information")

	flag.Usage = func() {
		flagSet := flag.CommandLine
		fmt.Printf("A command line tool for interacting with meshtastic radios\n")
		fmt.Printf("\n")
		fmt.Printf("USAGE\n")
		fmt.Printf("meshtastic-go -p <port> [COMMAND]\n")
		fmt.Printf("\n")
		fmt.Printf("COMMANDS\n")
		fmt.Printf("\n")
		order := []string{"port", "text", "info"}
		for _, name := range order {
			flag := flagSet.Lookup(name)
			fmt.Printf("-%s\t", flag.Name)
			fmt.Printf("  %s\n", flag.Usage)
		}
	}

	flag.Parse()

	if port == "" {
		flag.Usage()
		os.Exit(1)
	}

	radio := Radio{portNumber: port}
	radio.Init()
	defer radio.Close()

	if message != "" {
		radio.SendTextMessage(message)
	}

	if *infoPtr {
		responses, err := radio.GetRadioInfo()
		if err != nil {
			fmt.Println(err)
		}

		for _, response := range responses {
			fmt.Println(response)
		}
	}

}
