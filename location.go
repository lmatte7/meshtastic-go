package main

import (
	"github.com/urfave/cli/v2"
)

func setLocation(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	alt := int32(c.Int("alt"))
	err := radio.SetLocation(c.Float64("lat"), c.Float64("long"), alt)
	if err != nil {
		return err
	}

	return nil
}
