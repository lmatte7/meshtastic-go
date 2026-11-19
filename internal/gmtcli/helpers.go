package gmtcli

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lmatte7/gomesh"
	"github.com/urfave/cli/v2"
)

// portEnvVar names the environment variable consulted when --port is not
// given.
const portEnvVar = "MESHTASTIC_RADIO_DEV"

// darwinPortGlob matches the device files macOS creates for USB radios, both
// native USB boards (cu.usbmodem*) and USB-to-serial bridges (cu.usbserial*).
// The cu.* device is the one for outgoing connections.
const darwinPortGlob = "/dev/cu.usb*"

// These are variables so tests can stand in for the host.
var (
	goos      = runtime.GOOS
	globPorts = filepath.Glob
)

func getRadio(c *cli.Context) gomesh.Radio {
	port, err := resolvePort(c.String("port"))
	if err != nil {
		log.Fatal(err)
	}

	radio := gomesh.Radio{}
	if err := radio.Init(port); err != nil {
		log.Fatalf("Error setting radio port: %v", err)
	}

	return radio
}

// resolvePort decides which port to open. An explicit value (from --port or
// MESHTASTIC_RADIO_DEV, which urfave/cli folds into the flag) wins. Otherwise,
// on macOS, a single attached USB device is used; anything else is an error
// that tells the user what to do.
func resolvePort(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	if goos != "darwin" {
		return "", fmt.Errorf("no radio port given: pass --port or set %s", portEnvVar)
	}

	matches, err := globPorts(darwinPortGlob)
	if err != nil {
		return "", fmt.Errorf("looking for a radio: %w", err)
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", errors.New("no radio port given and no " + darwinPortGlob + " device found: pass --port or set " + portEnvVar)
	default:
		return "", fmt.Errorf("no radio port given and several devices match %s (%s): pass --port or set %s",
			darwinPortGlob, strings.Join(matches, ", "), portEnvVar)
	}
}
