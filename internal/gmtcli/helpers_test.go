package gmtcli

import (
	"errors"
	"strings"
	"testing"
)

func TestResolvePort(t *testing.T) {
	origGoos, origGlob := goos, globPorts
	defer func() { goos, globPorts = origGoos, origGlob }()

	fakeGlob := func(matches []string, err error) func(string) ([]string, error) {
		return func(pattern string) ([]string, error) {
			if pattern != darwinPortGlob {
				t.Errorf("glob pattern = %q, want %q", pattern, darwinPortGlob)
			}
			return matches, err
		}
	}

	cases := []struct {
		name     string
		goos     string
		explicit string
		matches  []string
		globErr  error
		want     string
		wantErr  string
	}{
		{"explicit port wins everywhere", "linux", "/dev/ttyUSB0", nil, nil, "/dev/ttyUSB0", ""},
		{"explicit port wins over a single device", "darwin", "192.168.0.42", []string{"/dev/cu.usbmodem1"}, nil, "192.168.0.42", ""},
		{"single usbmodem on macOS", "darwin", "", []string{"/dev/cu.usbmodem143101"}, nil, "/dev/cu.usbmodem143101", ""},
		{"single usbserial on macOS", "darwin", "", []string{"/dev/cu.usbserial-0001"}, nil, "/dev/cu.usbserial-0001", ""},
		{"several devices on macOS", "darwin", "", []string{"/dev/cu.usbmodem1", "/dev/cu.usbserial-0001"}, nil, "", "several devices match"},
		{"no device on macOS", "darwin", "", nil, nil, "", "no /dev/cu.usb* device found"},
		{"glob failure on macOS", "darwin", "", nil, errors.New("boom"), "", "looking for a radio"},
		{"linux needs an explicit port", "linux", "", []string{"/dev/cu.usbmodem1"}, nil, "", "pass --port or set MESHTASTIC_RADIO_DEV"},
	}
	for _, c := range cases {
		goos = c.goos
		globPorts = fakeGlob(c.matches, c.globErr)
		got, err := resolvePort(c.explicit)
		if c.wantErr == "" {
			if err != nil || got != c.want {
				t.Errorf("%s: resolvePort(%q) = %q, %v; want %q", c.name, c.explicit, got, err, c.want)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%s: resolvePort(%q) = %q, %v; want error containing %q", c.name, c.explicit, got, err, c.wantErr)
		}
	}
}
