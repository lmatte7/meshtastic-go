# Meshtastic Go

Meshtastic Go is a CLI for meshtastic devices built in Go. This tool provides an easy interface for meshtastic devices that can be used on Windows, Linux or Mac. The only requirements for this tool are the ESP32 [drivers](https://www.silabs.com/developers/usb-to-uart-bridge-vcp-drivers) if not already installed. Executables for different operating systems can be downloaded on the [releases](https://github.com/lmatte7/meshtastic-go/releases) page or can be built as needed with the standard Go tooling.

## Command syntax

A full list of commands can be viewed by running `--help`. Each command also has its own `--help` flag that provides more information on its subcommands and flags.

Every command requires the `--port` flag to be set to the port the radio is attached to. This can be set to a serial port (like `/dev/cu.SLAB_USBtoUART`) or an IP address depending on which communication method should be used to communicate with the radio. The CLI will automatically determine if TCP or serial communications should be used depending on what value is provided to `--port`.

```
NAME:
   meshtastic-go - Interface with meshtastic radios

USAGE:
   meshtastic-go [global options] command [command options] [arguments...]

VERSION:
   v0.2

AUTHOR:
   Lucas Matte <lmatte7@gmail.com>

COMMANDS:
   info      Show radio information
   message   Interact with radio messaging functionality
   channel   Update channel information
   prefs     Update user preferences
   location  Set location
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value, -p value  specify a port
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```

## Subcommands

### `info`

The `info` command displays information about the radio. By default all informaiton is shown, but information can be restricted by using subcommands. It is also possible to chain together subcommands, for example `meshtastic-go info rc` to only display radio and channel information

```
NAME:
   meshtastic-go info - Show radio information

USAGE:
   meshtastic-go info command [command options] [arguments...]

DESCRIPTION:
   Show node, preference and channel information for radio

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

### `message`

The `message` subcommand provides the ability to send messages and listen for new messages on the mesh. The `recv` subcommand won't show any previously received messages from the radio, but will wait and display new messages as they are received.

```
NAME:
   meshtastic-go message - Interact with radio messaging functionality

USAGE:
   meshtastic-go message command [command options] [arguments...]

DESCRIPTION:
   Send messages to other radios or wait for new messages

COMMANDS:
   send     Send a text message
   recv     Wait for new messages
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

### `stream`

The `stream` command is meant to be run as a long-lived subprocess by another program, for example a bot or web backend. It prints one JSON object per line (JSON Lines) to stdout and runs until it is interrupted or the radio link fails, in which case it exits non-zero.

```
meshtastic-go --port /dev/cu.usbmodem143101 stream
```

The first line has `"type": "ready"` and carries this radio's own node (`my_node`) and the configured channels. After that every text message the radio receives, direct or on a channel, is printed with `"type": "message"`, decrypted by the radio, with the sender (`from`), recipient (`to`, omitted for broadcasts) and `channel` resolved from the radio's node and channel database. `direct` is true for messages addressed to this radio, `broadcast` for channel messages.

```json
{"type":"message","received_at":"2026-11-19T04:12:08Z","id":1836540931,"rx_time":"2026-11-19T04:12:08Z","from":{"node_id":"!43b59594","node_number":1135973780,"longname":"DvdGmz01","shortname":"DG01","role":"CLIENT","hardware_model":"HELTEC_V3","is_licensed":false},"broadcast":true,"direct":false,"outgoing":false,"channel":{"index":0,"name":"","role":"PRIMARY"},"port":"TEXT_MESSAGE_APP","text":"hello mesh","snr_db":-1.75,"rssi_dbm":-110,"hop_limit":2,"hop_start":3,"hops_taken":1,"via_mqtt":false,"want_ack":false,"priority":"UNSET"}
```

With `--full`, every other packet is printed too with `"type": "packet"`: positions, telemetry, node info, routing and so on carry their payload as JSON under `payload` when the port is known, otherwise as `payload_base64`. Packets the radio could not decrypt have `"encrypted": true`.

The node and channel database is re-requested from the radio every five minutes, which also keeps the firmware forwarding packets to the connection.

#### Sending from the same process

With `--commands`, `stream` also reads JSON lines from stdin and sends each one as a text message, so a bot can reply without opening a second connection to the radio:

```json
{"id": "reply-17", "to": "!43b59594", "channel": 0, "text": "hello back"}
```

`to` is a node number, a `!xxxxxxxx` node id, or a node's long or short name (it must match exactly one node the radio knows); leave it out to broadcast on the channel. `channel` defaults to 0 and `id` is optional and echoed back. Each command produces one event, `{"type": "sent", ...}` with the resolved recipient and channel, or `{"type": "error", "error": "..."}` when the line was invalid or the radio refused it (for example a message over 240 bytes). When stdin is closed the stream exits with status 0.

The `sent` event carries the mesh `packet_id`, and the stream then reports what became of the message, each event echoing the command's `id`: `{"type": "queued"}` once the radio has accepted it for transmission, `{"type": "delivered"}` when it was acknowledged, or `{"type": "failed", "error": "MAX_RETRANSMIT"}` when the radio gave up, the queue rejected it, or nothing was heard for five minutes. For broadcasts, and whenever the acknowledgement comes from a node other than the recipient, `delivered` has `"implicit": true`: a neighbour rebroadcast the message, which proves it left this radio but not that anyone in particular received it.

Every message line also carries a `reply` object with two ready-made commands, so a bot answering something like `!weather` does not need to assemble addresses itself. `reply.direct` messages the sender privately and `reply.channel` answers publicly on the channel the message arrived on; copy one, add `text`, and write it back:

```json
"reply": {"direct": {"to": "!43b59594", "channel": 0}, "channel": {"channel": 0}}
```

### `config`

The `config` subcommand allows for different User Preferences (as defined in the [protobufs](https://github.com/lmatte7/goMesh/blob/6199a9555f0777b6f21456a1f5d1390bd324ba57/github.com/meshtastic/gomeshproto/radioconfig.pb.go#L422)) the be set and changed.

```
NAME:
   meshtastic-go config - Update radio config

USAGE:
   meshtastic-go config command [command options] [arguments...]

DESCRIPTION:
   Update radio config

COMMANDS:
   set      Set a user preference
   owner    Set the radio owner
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

### `location`

The `location` subcommand allows for the location to be manually set on the radio.

```
meshtastic-go location --help

NAME:
   meshtastic-go location - Set location

USAGE:
   meshtastic-go location command [command options] [arguments...]

DESCRIPTION:
   Manually set the GPS coordinates for the radio

COMMANDS:
   set      Set a location
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

### `channel`  
```
NAME:
   meshtastic-go channel - Update channel information

USAGE:
   meshtastic-go channel command [command options] [arguments...]

DESCRIPTION:
   Add, delete and update channel settings

COMMANDS:
   url      Change settings with a url
   add      Adds a channel
   delete   Deletes a channel
   set      Set a channel parameter
   options  Show channel options
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
   ```


### `reset`
```
NAME:
   meshtastic-go reset - Factory reset the radio

USAGE:
   reset - Factory reset the radio

DESCRIPTION:
   Reset the radio to default settings

OPTIONS:
   --help, -h  show help (default: false)
   ```

## Examples

Add a channel

```
meshtastic-go channel add --port=/dev/cu.SLAB_USBtoUART -i 2 -name test2
```

Update the Region

```
meshtastic-go config set -p /dev/cu.SLAB_USBtoUART -k Region -v 1
```

Set the radio location

```
meshtastic-go -p "192.168.0.42" location set --lat 310481775 --long -815817755 --alt 20
```

Send a message to all radios on the mesh

```
meshtastic-go --port "192.168.42.1" message send -m "test"
```

## Run smoke tests

To run the smoke tests, plug in a device, and run the following:

```
go build
cd smoke
go build
./smoke
```


## Feedback

This tool is still under development. Any issues or feedback can be submitted to the [issues](https://github.com/lmatte7/meshtastic-go/issues) page.
