package main

import (
	"github.com/urfave/cli/v2"
)

func setLocation(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	alt := int32(c.Int("alt"))
	return radio.SetLocation(int32(c.Int64("lat")), int32(c.Int64("long")), alt)
}
