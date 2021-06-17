# Meshtastic Go

Meshtastic Go is a CLI for meshtastic devices built in Go. This tool provides an easy interface for meshtastic devices that can be used on Windows, Linux or Mac. The only requirements for this tool are the ESP32 [drivers](https://www.silabs.com/developers/usb-to-uart-bridge-vcp-drivers) if not already installed. Executables for different operating systems can be downloaded on the [releases](https://github.com/lmatte7/meshtastic-go/releases) page or can be built as needed with the standard Go tooling.

## Command syntax

A full list of commands can be viewed by running `--help`. Each command also has its own `--help` flag that provides more information on its subcommands and flags. Every command requires the `--port` flag to be set with the current serial port the radio is attached to. This defaults to `/dev/cu.SLAB_USBtoUART`.

```
NAME:
   meshtastic-go - Interface with meshtastic radios

USAGE:
   meshtastic-go [global options] command [command options] [arguments...]

COMMANDS:
   info      Show radio information
   message   Interact with radio messaging functionality
   channel   Update channel information
   prefs     Update user preferences
   location  Set location
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value, -p value  specify a port (default: /dev/cu.SLAB_USBtoUART)
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```

## Subcommands

### `info`

The `info` command displays information about the radio. By default all informaiton is shown, but information can be restricted by using subcommands. It is also possible to chain together subcommands, for example `meshtastic-go info rc` to only display radio and channel information

```
meshtastic-go info --help

NAME:
   meshtastic-go info - Show radio information

USAGE:
   meshtastic-go info command [command options] [arguments...]

DESCRIPTION:
   Show node, preference and channel information for radio

COMMANDS:
   radio, r        Show radio information
   channels, c     Show all channel information
   nodes, n        Show all nodes on the mesh
   preferences, p  Show radio user preferences
   help, h         Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

### `message`

The `message` subcommand provides the ability to send messages and listen for new messages on the mesh. The `recv` subcommand won't show any previously recieved messages from the radio, but will wait and display new messages as they are recieved.

```
meshtastic-go message --help

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

### `prefs`

The `prefs` subcommand allows for different User Preferences (as defined in the [protobufs](https://github.com/lmatte7/goMesh/blob/6199a9555f0777b6f21456a1f5d1390bd324ba57/github.com/meshtastic/gomeshproto/radioconfig.pb.go#L422)) the be set and changed.

```
meshtastic-go prefs --help

NAME:
   meshtastic-go prefs - Update user preferences

USAGE:
   meshtastic-go prefs command [command options] [arguments...]

DESCRIPTION:
   Update user preferences

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

## Examples

Add a channel

```
meshtastic-go channel add -i 2 -name test2
```

Update the Region

```
meshtastic-go prefs set -k Region -v 1
```

Set the radio location

```
meshtastic-go location set --lat 310481775 --long -815817755 --alt 20
```

Send a message to all radios on the mesh

```
meshtastic-go message send -m "test"
```

## Feedback

This tool is still under development. Any issues or feedback can be submitted to the [issues](https://github.com/lmatte7/meshtastic-go/issues) page.
