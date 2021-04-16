package main

import (
	"fmt"
)

// option go_package = "github.com/lmatte7/meshtastic-go"

func main() {

	radio := Radio{portNumber: "/dev/cu.SLAB_USBtoUART"}

	radio.Init()

	defer radio.Close()

	responses, err := radio.GetRadioInfo()
	if err != nil {
		fmt.Println(err)
	}

	for _, response := range responses {
		fmt.Println(response)
	}

}
