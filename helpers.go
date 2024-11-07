package main

import (
	"log"

	"github.com/lmatte7/gomesh"
	"github.com/urfave/cli/v2"
)

func getRadio(c *cli.Context) gomesh.Radio {
	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		log.Fatalf("Error setting radio port: %v", err)
	}

	return radio
}
