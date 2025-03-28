package main

import "github.com/lmatte7/meshtastic-go/internal/gmtcli"

// option go_package = "github.com/lmatte7/meshtastic-go"

var VERSION = "unknown"
var COMMIT = "unknown"
var BUILDDATE = "unknown"

func main() {
	gmtcli.CLIEntry(VERSION, COMMIT, BUILDDATE)
}
