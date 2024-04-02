package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

var app string = "../meshtastic-go"

// TODO: could set the port as a command line arg
var port string = "/dev/cu.usbserial-0200674E"
var sleep_time time.Duration = (2 * time.Second)

func run_and_search(args []string, search string) {
	out, err := exec.Command(app, args...).Output()
	if err != nil {
		log.Fatal(err)
	}
	if search != "" {
		if !strings.Contains(string(out), search) {
			fmt.Printf("Did not find %s\n", search)
		}
	}
}

func smoke_info_r() {
	args := []string{"--port", port, "info", "r"}
	run_and_search(args, "Radio Settings:")
}

func smoke_info_c() {
	args := []string{"--port", port, "info", "c"}
	run_and_search(args, "Channel Settings:")
}

func smoke_info_n() {
	args := []string{"--port", port, "info", "n"}
	run_and_search(args, "Nodes")
}

func smoke_info_config() {
	args := []string{"--port", port, "config"}
	run_and_search(args, "Radio Config:")
}

func smoke_message_send() {
	args := []string{"--port", port, "message", "send", "-m", "test"}
	run_and_search(args, "")
}

func smoke_prefs_set() {
	// set the mqtt server to nothing
	args := []string{"--port", port, "config"}
	run_and_search(args, "Enabled                                 false")
	time.Sleep(sleep_time)
	args = []string{"--port", port, "config", "set", "-k", "Address", "-v", "foo"}
	run_and_search(args, "Address                                 foo")
}

func main() {
	smoke_info_r()
	smoke_info_c()
	smoke_info_n()
	smoke_info_config()
	smoke_message_send()
	smoke_prefs_set()
}
