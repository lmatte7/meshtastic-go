package main

import (
	"github.com/lmatte7/gomesh"
	"github.com/urfave/cli/v2"
)

func setLocation(c *cli.Context) error {
	radio := gomesh.Radio{}
	err := radio.Init(c.String("port"))
	if err != nil {
		return err
	}
	defer radio.Close()

	alt := int32(c.Int("alt"))
	err = radio.SetLocation(c.Float64("lat"), c.Float64("long"), alt)
	if err != nil {
		return err
	}

	return nil
}
