# Meshtastic Go
Meshtastic Go is a CLI for meshtastic devices built in Go. This tool provides an easy interface for meshtastic devices that can be used on Windows, Linux or Mac. The only requirements for this tool are the ESP32 [drivers](https://www.silabs.com/developers/usb-to-uart-bridge-vcp-drivers) if not already installed. Executables for different operating systems can be downloaded on the [releases](https://github.com/lmatte7/meshtastic-go/releases) page or can be built as needed with the standard Go tooling. 


## Command syntax
A full list of commands can be viewed by running `--help`
```
A command line tool for interacting with meshtastic radios

USAGE
meshtastic-go -p <port> [COMMAND]

COMMANDS

--port           --port=port The serial port for the radio (Required)
--info           Display radio information
--recv           Wait for new messages
--longSlow       Set long-range but slow channel
--shortFast      Set short-range but fast channel
--text           --text=message Send a text message
--setOwner       --setowner=owner Set the listed owner for the radio
--setKey         --setKey=key The key to set for a custom user preference option. Used with setValue
--setValue       --setValue=value The value to set for a custom user preference option. Used with setKey
```

## Feedback
This tool is still under development. Any issues or feedback can be submitted to the [issues](https://github.com/lmatte7/meshtastic-go/issues) page.