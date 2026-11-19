# Meshtastic Go

Meshtastic Go is a CLI for [Meshtastic](https://meshtastic.org/) devices built
in Go. It runs on Windows, Linux and macOS and talks to a radio over a serial
port or TCP. ESP32 boards with a USB-to-serial bridge may need the
[Silicon Labs drivers](https://www.silabs.com/developers/usb-to-uart-bridge-vcp-drivers);
boards with native USB (nRF52 devices such as the RAK4631 or T-Echo) need no
driver.

Executables can be downloaded from the
[releases](https://github.com/lmatte7/meshtastic-go/releases) page or built
from source with `make build`.

The protobuf bindings are generated from the current
[Meshtastic protobufs](https://github.com/meshtastic/protobufs) release (see
[Protobufs](#protobufs) below), so the tool understands the fields reported
by current firmware.

## Command syntax

A full list of commands can be viewed by running `--help`. Each command also
has its own `--help` flag that provides more information on its subcommands
and flags.

Every command requires the `--port` flag to be set to the port the radio is
attached to. This can be a serial port (for example `/dev/cu.usbmodem143101`
or `/dev/cu.SLAB_USBtoUART`) or an IP address, in which case TCP is used.

```
NAME:
   meshtastic-go - Interface with meshtastic radios

USAGE:
   meshtastic-go [global options] command [command options] [arguments...]

COMMANDS:
   info      Show radio information
   message   Interact with radio messaging functionality
   channel   Update channel information
   config    Update radio config
   location  Set location
   reset     Factory reset the radio
   modem     Set the modem mode
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value, -p value  specify a port
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```

## Subcommands

### `info`

The `info` command displays information about the radio. By default all
information is shown, but it can be restricted with subcommands. The `--json`
flag on `info` prints the raw radio, channel and node packets as JSON.

```
NAME:
   meshtastic-go info - Show radio information

USAGE:
   meshtastic-go info command [command options] [arguments...]

COMMANDS:
   radio, r     Show radio information
   channels, c  Show all channel information
   nodes, n     Show all nodes on the mesh
   position, p  Show position settings
   metrics, m   Display node metrics
   help, h      Shows a list of commands or help for one command

OPTIONS:
   --json      Output data in JSON (default: false)
   --help, -h  show help (default: false)
```

`info nodes` prints a table of every node the radio knows about. With
`--json` it prints a list of node objects with resolved names and units,
suitable for scripting:

```
meshtastic-go --port /dev/cu.usbmodem143101 info nodes --json
```

```json
[
  {
    "node_id": "!43b59594",
    "node_number": 1135973780,
    "longname": "DvdGmz01",
    "shortname": "DG01",
    "altitude_meters": 471,
    "latitude_degrees": 36.1824256,
    "longitude_degrees": -115.1926272,
    "mac_address": "48:ca:43:b5:95:94",
    "public_key": "+wbec92kaoh/kIxrntNL75cAPf/r5iOfdoOrbacohBs=",
    "is_licensed": false,
    "role": "CLIENT",
    "hardware_model": "HELTEC_V3",
    "battery_level_percent": 101,
    "battery_volts": 4.234000205993652,
    "channel_utilization_percent": 12.451666831970215,
    "air_utilization_transmit_percent": 0.11852778494358063,
    "uptime_seconds": 121,
    "snr_db": -1.75,
    "hops_away": 1,
    "via_mqtt": false,
    "is_favorite": false,
    "is_ignored": false,
    "last_heard": "2026-11-14T17:46:01Z",
    "last_heard_seconds_ago": 376594,
    "last_heard_human": "4d8h36m"
  }
]
```

Fields the radio did not report are omitted.

### `message`

The `message` command sends messages and listens for new messages on the
mesh. `receive` does not show previously received messages, but waits and
displays new text messages as they arrive. With `--json` each message is
printed as one JSON object per line, and `--exit` stops after the first
message.

```
NAME:
   meshtastic-go message - Interact with radio messaging functionality

USAGE:
   meshtastic-go message command [command options] [arguments...]

COMMANDS:
   send     Send a text message
   receive  Wait for new messages
   help, h  Shows a list of commands or help for one command
```

### `config`

The `config` command shows the radio and module configuration, and `config
set` changes a single value by its field name as shown in the output.

```
NAME:
   meshtastic-go config - Update radio config

USAGE:
   meshtastic-go config command [command options] [arguments...]

COMMANDS:
   set      Set a user preference
   owner    Set the radio owner
   help, h  Shows a list of commands or help for one command
```

### `location`

The `location` command sets a fixed position on the radio. Coordinates are
integers in units of 1e-7 degrees, altitude is in metres.

```
NAME:
   meshtastic-go location - Set location

USAGE:
   meshtastic-go location command [command options] [arguments...]

COMMANDS:
   set      Set a location
   help, h  Shows a list of commands or help for one command
```

### `channel`

```
NAME:
   meshtastic-go channel - Update channel information

USAGE:
   meshtastic-go channel command [command options] [arguments...]

COMMANDS:
   url      Change settings with a url
   add      Adds a channel
   delete   Deletes a channel
   set      Set a channel parameter
   options  Show channel options
   help, h  Shows a list of commands or help for one command
```

### `modem`

```
NAME:
   meshtastic-go modem - Set the modem mode

USAGE:
   meshtastic-go modem command [command options] [arguments...]

COMMANDS:
   options  show modem options
   set      set a modem option
   help, h  Shows a list of commands or help for one command
```

### `reset`

`reset` performs a factory reset of the radio. There is no confirmation.

## Examples

Add a channel

```
meshtastic-go --port /dev/cu.SLAB_USBtoUART channel add -i 2 -n test2
```

Update the region

```
meshtastic-go -p /dev/cu.SLAB_USBtoUART config set -k Region -v 1
```

Set the radio location

```
meshtastic-go -p 192.168.0.42 location set --lat 310481775 --long -815817755 --alt 20
```

Send a message to all radios on the mesh

```
meshtastic-go --port 192.168.42.1 message send -m "test"
```

Follow incoming text messages as JSON

```
meshtastic-go --port /dev/cu.usbmodem143101 message receive --json
```

## Building and testing

The `Makefile` wraps the usual steps:

| Target           | What it does                                           |
| ---------------- | ------------------------------------------------------ |
| `make build`     | runs the tests, then cross-compiles into `builds/`     |
| `make test`      | runs the unit tests                                    |
| `make fmt`       | formats the Go sources with `gofmt`                    |
| `make fmt-check` | fails if any file is not `gofmt`-clean                 |
| `make lint`      | runs `golangci-lint`                                   |
| `make proto`     | regenerates the protobuf bindings (see below)          |
| `make clean`     | removes `builds/`                                      |

The radio library in `internal/gomesh` has a hardware test suite that is
skipped unless a port is given:

```
go test ./internal/gomesh -args -port=/dev/cu.usbmodem143101
```

Those tests change settings on the attached radio (owner name, channels,
modem preset, location), so only run them against a radio you can
reconfigure.

### Protobufs

`internal/gomesh/gomeshproto` is generated from
[meshtastic/protobufs](https://github.com/meshtastic/protobufs) by
`genproto.sh`, which pins the upstream tag it generates from. To update to a
newer release, change `PROTOBUFS_TAG` in the script and run `make proto`
with `protoc` and `protoc-gen-go` installed, then fix anything the compiler
complains about.

## Run smoke tests

To run the smoke tests, plug in a device, and run the following:

```
go build ./cmd/meshtastic-go
cd smoke
go build
./smoke
```

## Feedback

This tool is still under development. Any issues or feedback can be submitted
to the [issues](https://github.com/lmatte7/meshtastic-go/issues) page.
