# Meshtastic Go

Meshtastic Go is a CLI for [Meshtastic](https://meshtastic.org/) devices built
in Go. It runs on Windows, Linux and macOS and talks to a radio over a serial
port or TCP. ESP32 boards with a USB-to-serial bridge may need the
[Silicon Labs drivers](https://www.silabs.com/developers/usb-to-uart-bridge-vcp-drivers);
boards with native USB (nRF52 devices such as the RAK4631 or T-Echo) need no
driver.

Besides the usual commands for inspecting and configuring a radio, it has a
`stream` command that turns the radio into a subprocess any program can talk
to with JSON lines, which is how the example bots under `examples/` are built.

Executables can be downloaded from the
[releases](https://github.com/lmatte7/meshtastic-go/releases) page or built
from source with `make build`.

## Quickstart

Setting up a radio fresh out of the box takes a handful of commands. Plug it
in, tell the CLI where it is, and set the region (required before the radio
will transmit), your names, the modem preset, the hop limit and the role:

```
export MESHTASTIC_RADIO_DEV=/dev/cu.usbmodem143101   # on macOS, optional with one radio attached

meshtastic-go config set -k Region -v US              # or EU_868, EU_433, ANZ, ... (see "config")
meshtastic-go config owner -n "Alice's T-Echo" -s ALCE
meshtastic-go modem set -o lf                         # LONG_FAST, the default most meshes use
meshtastic-go config set -k HopLimit -v 3
meshtastic-go config set -k Role -v CLIENT            # or CLIENT_MUTE, ROUTER, ...
meshtastic-go config set -k GpsMode -v ENABLED        # if the radio has a GPS
meshtastic-go channel url -u 'https://meshtastic.org/e/#...'   # to join a private channel

meshtastic-go info                                    # check the result
```

The radio reboots to apply some of these, so a command may take a few
seconds while the next one waits for it to come back. Enum values are given
by name as shown in `config` output; `config` lists every field that can be
set.

The protobuf bindings are generated from the current
[Meshtastic protobufs](https://github.com/meshtastic/protobufs) release (see
[Protobufs](#protobufs) below), so the tool understands the fields reported
by current firmware.

## Command syntax

A full list of commands can be viewed by running `--help`. Each command also
has its own `--help` flag that provides more information on its subcommands
and flags.

Every command needs to know the port the radio is attached to: a serial port
(like `/dev/cu.usbmodem143101`) or an IP address, in which case TCP is used.
The port is taken from, in order:

1. the `--port` flag;
2. the `MESHTASTIC_RADIO_DEV` environment variable;
3. on macOS, the only `/dev/cu.usb*` device present, if there is exactly one.

Setting the environment variable once is the convenient way to work with a
radio, and the examples below assume it:

```
export MESHTASTIC_RADIO_DEV=/dev/cu.usbmodem143101
```

If none of the three yields a port, or several USB devices are attached, the
command exits with a message saying so.

```
NAME:
   meshtastic-go - Interface with meshtastic radios

USAGE:
   meshtastic-go [global options] command [command options] [arguments...]

COMMANDS:
   info        Show radio information
   message     Interact with radio messaging functionality
   stream      Stream received messages to stdout as JSON lines
   channel     Update channel information
   config      Update radio config
   location    Set location
   locate      Show where the radio is
   traceroute  Trace the route to a node
   nodedb      Manage the radio's node database
   reboot      Reboot the radio
   reset       Factory reset the radio
   modem       Set the modem mode
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value, -p value  serial port or IP address of the radio; on macOS defaults to the only /dev/cu.usb* device if there is exactly one [$MESHTASTIC_RADIO_DEV]
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
meshtastic-go info nodes --json
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
        "battery_volts": 4.234,
        "channel_utilization_percent": 12.45,
        "air_utilization_transmit_percent": 0.12,
        "uptime_seconds": 121,
        "snr_db": -1.75,
        "hops_away": 1,
        "via_mqtt": false,
        "is_favorite": false,
        "is_ignored": false,
        "last_heard": "2026-11-19T05:46:01Z",
        "last_heard_seconds_ago": 94,
        "last_heard_human": "1m34s"
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

### `stream`

The `stream` command is meant to be run as a long-lived subprocess by another
program, for example a bot or web backend. It prints one JSON object per line
(JSON Lines) to stdout and runs until it is interrupted or the radio link
fails, in which case it exits non-zero. Pipe it through `jq` for a
human-readable view.

```
meshtastic-go stream
```

The first line has `"type": "ready"` and carries this radio's own node
(`my_node`) and the configured channels. After that every text message the
radio receives, direct or on a channel, is printed with `"type": "message"`,
decrypted by the radio, with the sender (`from`), recipient (`to`, omitted
for broadcasts) and `channel` resolved from the radio's node and channel
database. `direct` is true for messages addressed to this radio, `broadcast`
for channel messages.

The output is one object per line; it is shown formatted here for
readability, which means exactly the same thing.

```json
{
    "type": "message",
    "received_at": "2026-11-19T04:12:08Z",
    "id": 1836540931,
    "from": {
        "node_id": "!43b59594",
        "node_number": 1135973780,
        "longname": "DvdGmz01",
        "shortname": "DG01",
        "role": "CLIENT",
        "hardware_model": "HELTEC_V3",
        "is_licensed": false
    },
    "broadcast": true,
    "direct": false,
    "outgoing": false,
    "channel": {
        "index": 0,
        "name": "",
        "role": "PRIMARY"
    },
    "port": "TEXT_MESSAGE_APP",
    "text": "hello mesh",
    "reply": {
        "direct": {
            "to": "!43b59594",
            "channel": 0
        },
        "channel": {
            "channel": 0
        }
    },
    "snr_db": -1.75,
    "rssi_dbm": -110,
    "hop_limit": 2,
    "hop_start": 3,
    "hops_taken": 1,
    "via_mqtt": false,
    "want_ack": false,
    "priority": "UNSET"
}
```

With `--full`, every other packet is printed too with `"type": "packet"`:
positions, telemetry, node info, routing and so on carry their payload as
JSON under `payload` when the port is known (with readings rounded to a
sensible precision), otherwise as `payload_base64`. Packets the radio could
not decrypt have `"encrypted": true`.

Timestamps in the stream are the stream's own: `received_at` on every line
is when the packet arrived, and in `nodes` events the `last_heard` fields
say when the stream last received a packet from that node during this run.
Nodes it has not heard yet carry no last-heard fields. The radio's own
timestamps are not reported, because a radio without a time source (no GPS
fix, no phone) stamps with a clock that can be days off.

The node and channel database is re-requested from the radio every five
minutes, which also keeps the firmware forwarding packets to the connection.

#### Sending from the same process

With `--commands`, `stream` also reads JSON lines from stdin and sends each
one as a text message, so a bot can reply without opening a second
connection to the radio:

```json
{
    "id": "reply-17",
    "to": "!43b59594",
    "channel": 0,
    "text": "hello back"
}
```

`to` is a node number, a `!xxxxxxxx` node id, or a node's long or short name
(it must match exactly one node the radio knows); leave it out to broadcast
on the channel. `channel` defaults to 0 and `id` is optional and echoed back.
Each command produces one event, `{"type": "sent", ...}` with the resolved
recipient and channel, or `{"type": "error", "error": "..."}` when the line
was invalid or the radio refused it (for example a message over 237 bytes).
When stdin is closed the stream exits with status 0.

The `sent` event carries the mesh `packet_id`, and the stream then reports
what became of the message, each event echoing the command's `id`:
`{"type": "queued"}` once the radio has accepted it for transmission,
`{"type": "delivered"}` when it was acknowledged, or
`{"type": "failed", "error": "MAX_RETRANSMIT"}` when the radio gave up, the
queue rejected it, or nothing was heard for five minutes. For broadcasts,
and whenever the acknowledgement comes from a node other than the recipient,
`delivered` has `"implicit": true`: a neighbour rebroadcast the message,
which proves it left this radio but not that anyone in particular received
it.

Every message line also carries a `reply` object with two ready-made
commands, so a bot answering something like `!weather` does not need to
assemble addresses itself. `reply.direct` messages the sender privately and
`reply.channel` answers publicly on the channel the message arrived on; copy
one, add `text`, and write it back:

```json
"reply": {
    "direct": {
        "to": "!43b59594",
        "channel": 0
    },
    "channel": {
        "channel": 0
    }
}
```

A command line may also ask the stream for its cached node list instead of
sending:

```json
{
    "id": "n1",
    "command": "nodes"
}
```

It is answered with a `nodes` event carrying the same node records
`info nodes --json` prints, including `hops_away`, `via_mqtt` and `snr_db`,
so a bot can reason about the mesh without opening a second connection to
the radio:

```json
{
    "type": "nodes",
    "at": "2026-11-19T04:12:09Z",
    "id": "n1",
    "my_node": {
        "node_id": "!534c8a43",
        "node_number": 1397525059,
        "longname": "Meshtastic 8a43",
        "shortname": "8a43",
        "hardware_model": "T_ECHO"
    },
    "nodes": [
        {
            "node_id": "!43b59594",
            "node_number": 1135973780,
            "longname": "DvdGmz01",
            "shortname": "DG01",
            "latitude_degrees": 36.1824256,
            "longitude_degrees": -115.1926272,
            "hops_away": 1,
            "via_mqtt": false,
            "snr_db": -1.75,
            "last_heard": "2026-11-19T04:10:51Z",
            "last_heard_seconds_ago": 78,
            "last_heard_human": "1m18s"
        }
    ]
}
```

#### Example bots

Each directory under `examples/` is a complete bot in standard-library Go.
They run `meshtastic-go stream --commands` as a subprocess, read its lines
and write commands back, so the same pattern works from any language.

`examples/pingbot` answers any `!ping` with `pong` on the channel it arrived
on (or directly for a direct message), and logs the `sent`, `queued`,
`delivered` and `failed` events for its replies:

```
go run ./examples/pingbot
```

`examples/weatherbot` answers `!weather` with the current weather from
[wttr.in](https://wttr.in) for where the radio is. It uses the radio's GPS
position when it has one; otherwise it asks the stream for the node list and
takes the mean position of nodes heard over the air (not via MQTT) within
`-max-hops` hops (default 2), preferring nodes the stream has heard within
`-max-age` (default 24h) and falling back to all of them when fewer than
three have been heard yet. `!weather <place>` looks up a named place
instead. Replies go back the way the request came and contain the weather
only; how the position was determined is logged, never sent to the mesh. The bot is throttled as a
whole: at most one lookup and one reply per `-interval` (default 5m) no
matter how many nodes ask, so a hostile mesh cannot use it to flood wttr.in
or the channel. The first sender turned away in a period is told the cooldown
by direct message; further requests in that period get no answer at all:

```
go run ./examples/weatherbot
```

The bots find the radio the same way the CLI does (`-port`,
`MESHTASTIC_RADIO_DEV`, or the only attached USB radio on macOS) and expect
`meshtastic-go` on the `PATH` (`-bin` points elsewhere). `-debug` logs every
line received from the stream, every command sent, and the reason a message
was ignored, which is the first thing to turn on when a bot does not answer.

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

### `locate`

`locate` prints the radio's GPS position if it has a fix. Without one it
estimates the position from the nodes the radio heard directly (zero hops,
not via MQTT) in the last 15 minutes that report a position, as the mean of
their positions weighted by received signal power (10^(SNR/10)), so a strong
nearby neighbour counts for far more than a faint distant one. `--json`
prints the position and the contributing nodes with their weights.

```
meshtastic-go locate
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

### `traceroute`

`traceroute <node>` asks the mesh for the route to a node and waits up to
`--timeout` (default 30s) for the reply. The node is a node number, a
`!xxxxxxxx` id, or a long or short name. The output lists each relay the
request passed through on its way out and the reply on its way back, with the
signal-to-noise ratio each receiving node measured; `--json` gives the same
as data. A routing error from the mesh (for example no route, or retransmits
exhausted) ends the wait early with that reason.

```
meshtastic-go traceroute DvdGmz01
meshtastic-go traceroute --timeout 2m !43b59594
```

### `nodedb`

The radio keeps a fixed-size table of the nodes it has heard (80 entries on
nRF52 boards such as the T-Echo, 100 on larger ones). It can fill up with
nodes from days ago that are never heard again, at which point recently
heard neighbours may not get an entry and `info nodes` and `locate` do not
see them. `nodedb list` prints the table (the same as `info nodes`, with
`--json` for node records), and `nodedb reset` asks the radio to forget every
node; it rebuilds the table from the nodes it hears from then on. There is no
confirmation.

```
meshtastic-go nodedb list
meshtastic-go nodedb reset
```

### `reboot`

`reboot` restarts the radio; settings and the node database are kept.

### `reset`

`reset` performs a factory reset of the radio. There is no confirmation.

## Examples

With the radio's port exported once:

```
export MESHTASTIC_RADIO_DEV=/dev/cu.usbmodem143101
```

Add a channel

```
meshtastic-go channel add -i 2 -n test2
```

Update the region

```
meshtastic-go config set -k Region -v 1
```

Set the radio location

```
meshtastic-go location set --lat 310481775 --long -815817755 --alt 20
```

Send a message to all radios on the mesh

```
meshtastic-go message send -m "test"
```

Follow incoming text messages as JSON

```
meshtastic-go message receive --json
```

A radio reached over TCP instead of serial:

```
meshtastic-go --port 192.168.42.1 info nodes
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

The radio library in `pkg/gomesh` has a hardware test suite that is
skipped unless a port is given:

```
go test ./pkg/gomesh -args -port=/dev/cu.usbmodem143101
```

Those tests change settings on the attached radio (owner name, channels,
modem preset, location), so only run them against a radio you can
reconfigure.

### Protobufs

`pkg/gomesh/gomeshproto` is generated from
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
