package gmtcli

import (
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

// CLIEntry starts the CLI and determines flags
func CLIEntry(version, commit, builddate string) {

	app := &cli.App{
		Name:    "meshtastic-go",
		Version: version + " (" + commit + ") on " + builddate,
		Authors: []*cli.Author{
			{
				Name: "Lucas Matte and contributors",
			},
		},
		Usage: "Interface with meshtastic radios",
		Commands: []*cli.Command{
			{
				Name:        "info",
				Usage:       "Show radio information",
				UsageText:   "info [command] - Show radio information",
				Description: "Show node, preference and channel information for radio",
				ArgsUsage:   "",
				Action:      showAllRadioInfo,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "json",
						Usage:    "Output data in JSON",
						Required: false,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:    "radio",
						Aliases: []string{"r"},
						Usage:   "Show radio information",
						Action:  showRadioInfo,
					},
					{
						Name:    "channels",
						Aliases: []string{"c"},
						Usage:   "Show all channel information",
						Action:  showChannelInfo,
					},
					{
						Name:    "nodes",
						Aliases: []string{"n"},
						Usage:   "Show all nodes on the mesh",
						Action:  showNodeInfo,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:     "json",
								Usage:    "Output nodes in JSON",
								Required: false,
							},
						},
					},
					{
						Name:    "position",
						Aliases: []string{"p"},
						Usage:   "Show position settings",
						Action:  showPositionInfo,
					},
					{
						Name:    "metrics",
						Aliases: []string{"m"},
						Usage:   "Display node metrics",
						Action:  showMetricInfo,
					},
				},
			},
			{
				Name:        "message",
				Usage:       "Interact with radio messaging functionality",
				UsageText:   "message [command]",
				Description: "Send messages to other radios or wait for new messages",
				ArgsUsage:   "",
				Subcommands: []*cli.Command{
					{
						Name:        "send",
						Usage:       "Send a text message",
						UsageText:   "text - Sends a text message to a node",
						Description: "Sends a text message to a Node, or to all nodes if no address is provided",
						Action:      sendText,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "message",
								Aliases:  []string{"m"},
								Usage:    "Message to send, must be under 237 characters",
								Required: true,
							},
							&cli.Int64Flag{
								Name:    "to",
								Aliases: []string{"t"},
								Usage:   "Address to send to. Leave blank for broadcast",
								Value:   0,
							},
							&cli.Int64Flag{
								Name:    "channel",
								Aliases: []string{"c"},
								Usage:   "Channel to send the message on",
								Value:   0,
							},
						},
					},
					{
						Name:        "receive",
						Usage:       "Wait for new messages",
						Description: "Waits for new messages and displays them as received until cancelled. Only shows messages on TEXT_MESSAGE port. Doesn't work over TCP",
						Action:      getReceivedMessages,
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:    "channel",
								Aliases: []string{"c"},
								Usage:   "Channel to listen for messages on",
								Value:   0,
							},
							&cli.BoolFlag{
								Name:     "json",
								Usage:    "Output messages in JSON with a newline between each message",
								Required: false,
							},
							&cli.BoolFlag{
								Name:     "exit",
								Aliases:  []string{"e"},
								Usage:    "Exit after receiving a message from the mesh",
								Required: false,
								Value:    false,
							},
						},
					},
				},
			},
			{
				Name:        "stream",
				Usage:       "Stream received messages to stdout as JSON lines",
				UsageText:   "stream [--full] - stream messages until interrupted",
				Description: "Prints one JSON object per line for every text message (direct or channel, decrypted by the radio) until interrupted, after an initial \"ready\" line with this radio's node and channels. Every line carries the sender, recipient and channel details. With --full every other packet type is streamed as well. Intended to be run as a subprocess by other programs.",
				ArgsUsage:   "",
				Action:      streamMessages,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "full",
						Usage: "Also stream positions, telemetry, node info, routing and undecryptable packets",
					},
					&cli.BoolFlag{
						Name:  "commands",
						Usage: "Read JSON lines from stdin ({\"to\": node, \"channel\": index, \"text\": \"...\"}) and send them as text messages; exit when stdin closes",
					},
				},
			},
			{
				Name:        "channel",
				Usage:       "Update channel information",
				UsageText:   "channel [command] - Update channel parameters",
				Description: "Add, delete and update channel settings",
				ArgsUsage:   "",
				Action:      showChannelInfo,
				Subcommands: []*cli.Command{
					{
						Name:        "url",
						Usage:       "Change settings with a url",
						UsageText:   "url - change settings with url",
						Description: "Set channel settings on radio using a meshtastic URL",
						Action:      setUrl,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "url",
								Aliases:  []string{"u"},
								Usage:    "Meshtastic channel URL to use",
								Required: true,
							},
						},
					},
					{
						Name:        "add",
						Usage:       "Adds a channel",
						UsageText:   "add - Add a channel to the radio",
						Description: "Add a channel to the radio with a random PSK",
						Action:      addChannel,
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "index",
								Aliases:     []string{"i"},
								Usage:       "Index for the channel to be added. If a channel is added at 0 it will become the Primary channel",
								Value:       1,
								DefaultText: "1",
								Required:    true,
							},
							&cli.StringFlag{
								Name:     "name",
								Aliases:  []string{"n"},
								Usage:    "Name of the chanel",
								Required: true,
							},
						},
					},
					{
						Name:        "delete",
						Usage:       "Deletes a channel",
						UsageText:   "delete - Delete a channel to the radio",
						Description: "Delete a channel from the radio. Cannot delete a PRIMARY channel",
						Action:      deleteChannel,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "index",
								Aliases:  []string{"i"},
								Usage:    "Index for the channel to be added. If a channel is added at 0 it will become the Primary channel",
								Required: true,
							},
						},
					},
					{
						Name:        "set",
						Usage:       "Set a channel parameter",
						Description: "Sets channel parameters for the specified channel index",
						Action:      setChannel,
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "index",
								Aliases:     []string{"i"},
								Usage:       "Index for the channel to be added. If a channel is added at 0 it will become the Primary channel",
								Required:    true,
								Value:       1,
								DefaultText: "1",
							},
							&cli.StringFlag{
								Name:     "key",
								Aliases:  []string{"k"},
								Usage:    "Key of the channel parameter to be changed",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "value",
								Aliases:  []string{"v"},
								Usage:    "Value of the parameter",
								Required: true,
							},
						},
					},
					{
						Name:        "options",
						Usage:       "Show channel options",
						Description: "Show all avaible channel options that can be set",
						Action:      showChannelOptions,
					},
				},
			},
			{
				Name:        "config",
				Usage:       "Update radio config",
				UsageText:   "config [command] - Update radio config",
				Description: "Update radio config",
				ArgsUsage:   "",
				Action:      showRadioConfig,
				Subcommands: []*cli.Command{
					{
						Name:        "set",
						Usage:       "Set a config value",
						Description: "Sets one radio or module config value by the field name shown in \"config\" output. Enum values are given by name (for example Region US, Role CLIENT, ModemPreset LONG_FAST).",
						Action:      setConfig,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "key",
								Aliases:  []string{"k"},
								Usage:    "Field name as shown by \"config\", for example Region or HopLimit",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "value",
								Aliases:  []string{"v"},
								Usage:    "The value: a number, true/false, text, or an enum name",
								Required: true,
							},
						},
					},
					{
						Name:        "owner",
						Usage:       "Set the radio's long and short name",
						Description: "Sets the names other nodes see for this radio. The short name is at most four characters and defaults to the first four of the long name.",
						Action:      setOwner,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Aliases:  []string{"n"},
								Usage:    "The long name",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "short",
								Aliases: []string{"s"},
								Usage:   "The short name, up to 4 characters",
							},
						},
					},
				},
			},
			{
				Name:        "location",
				Usage:       "Set location",
				UsageText:   "location [command] - Set location",
				Description: "Manually set GPS coordinates using int values for lat, long, and alt",
				ArgsUsage:   "",
				Subcommands: []*cli.Command{
					{
						Name:        "set",
						Usage:       "Set a location",
						Description: "Manually set GPS coordinates using int values for lat, long, and alt",
						Action:      setLocation,
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:     "lat",
								Usage:    "Latitude",
								Required: true,
							},
							&cli.Int64Flag{
								Name:     "long",
								Usage:    "Longitude",
								Required: true,
							},
							&cli.IntFlag{
								Name:     "alt",
								Usage:    "Altitude",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:        "locate",
				Usage:       "Show where the radio is",
				UsageText:   "locate [--json] - show the radio's GPS fix, or estimate its position",
				Description: "Prints the radio's GPS position if it has a fix. Otherwise estimates its position from the nodes it heard directly (not via MQTT) in the last 15 minutes that report a position, weighting each by its signal strength.",
				ArgsUsage:   "",
				Action:      locateRadio,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "json",
						Usage: "Output the position and contributing nodes as JSON",
					},
				},
			},
			{
				Name:        "traceroute",
				Usage:       "Trace the route to a node",
				UsageText:   "traceroute [--timeout 30s] [--json] <node>",
				Description: "Sends a traceroute request to a node, given as a node number, a !xxxxxxxx id, or a long or short name, and prints the relays and per-hop signal-to-noise ratios the request and the reply travelled through.",
				ArgsUsage:   "<node>",
				Action:      traceroute,
				Flags: []cli.Flag{
					&cli.DurationFlag{
						Name:  "timeout",
						Usage: "How long to wait for the reply",
						Value: 30 * time.Second,
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "Output the route as JSON",
					},
				},
			},
			{
				Name:        "nodedb",
				Usage:       "Manage the radio's node database",
				UsageText:   "nodedb [command]",
				Description: "The radio keeps a table of the nodes it has heard, which has a fixed size and can fill up with stale entries.",
				ArgsUsage:   "",
				Subcommands: []*cli.Command{
					{
						Name:        "list",
						Usage:       "List the nodes in the database",
						Description: "Prints every node the radio knows about, as a table or with --json as node records.",
						Action:      showNodeInfo,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "json",
								Usage: "Output nodes in JSON",
							},
						},
					},
					{
						Name:        "reset",
						Usage:       "Clear the node database",
						Description: "Asks the radio to forget every node it knows about. It rebuilds the table from the nodes it hears from then on. There is no confirmation.",
						Action:      resetNodeDB,
					},
				},
			},
			{
				Name:        "reset",
				Usage:       "Factory reset the radio",
				UsageText:   "reset - Factory reset the radio",
				Description: "Reset the radio to default settings",
				ArgsUsage:   "",
				Action:      factoryResetRadio,
			},
			{
				Name:        "modem",
				Usage:       "Set the modem mode",
				UsageText:   "modem [command] - set modem mode",
				Description: "Set the modem mode for the radio",
				ArgsUsage:   "",
				Action:      showModemOptions,
				Subcommands: []*cli.Command{
					{
						Name:        "options",
						Usage:       "show modem options",
						Description: "Show available modem options",
						Action:      showModemOptions,
					},
					{
						Name:        "set",
						Usage:       "set a modem option",
						Description: "Set the modem",
						Action:      setModemOption,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "option",
								Aliases:  []string{"o"},
								Usage:    "Modem preset short code or name, see 'modem options'",
								Required: true,
							},
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "serial port or IP address of the radio; on macOS defaults to the only /dev/cu.usb* device if there is exactly one",
				EnvVars: []string{"MESHTASTIC_RADIO_DEV"},
			},
		},
	}

	app.Run(os.Args)

}
