package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

// Init starts the CLI and determines flags
func Init() {

	app := &cli.App{
		Name:    "meshtastic-go",
		Version: "v0.2",
		Authors: []*cli.Author{
			{
				Name:  "Lucas Matte",
				Email: "lmatte7@gmail.com",
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
					},
					{
						Name:    "preferences",
						Aliases: []string{"p"},
						Usage:   "Show radio user preferences",
						Action:  showRadioPreferences,
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
								Usage:    "Message to send, must be under 240 characters",
								Required: true,
							},
							&cli.Int64Flag{
								Name:    "to",
								Aliases: []string{"t"},
								Usage:   "Address to send to. Leave blank for broadcast",
								Value:   0,
							},
						},
					},
					{
						Name:        "recv",
						Usage:       "Wait for new messages",
						Description: "Waits for new messages and displays them as recieved until cancelled. Only shows messages on TEXT_MESSAGE port",
						Action:      getRecievedMessages,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:     "exit",
								Aliases:  []string{"e"},
								Usage:    "Exit after recieving a message from the mesh",
								Required: true,
							},
						},
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
				},
			},
			{
				Name:        "prefs",
				Usage:       "Update user preferences",
				UsageText:   "prefs [command] - Update user preferences",
				Description: "Update user preferences",
				ArgsUsage:   "",
				Action:      showRadioPreferences,
				Subcommands: []*cli.Command{
					{
						Name:        "set",
						Usage:       "Set a user preference",
						Description: "Sets a user preference using the provided key/value combination",
						Action:      setPref,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "key",
								Aliases:  []string{"k"},
								Usage:    "Key of the user preferences to be changed",
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
						Name:        "owner",
						Usage:       "Set the radio owner",
						Description: "Sets the owner of the radio that is sent out over the network",
						Action:      setOwner,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Aliases:  []string{"n"},
								Usage:    "The owner name",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:        "location",
				Usage:       "Set location",
				UsageText:   "location [command] - Set location",
				Description: "Manually set the GPS coordinates for the radio",
				ArgsUsage:   "",
				Subcommands: []*cli.Command{
					{
						Name:        "set",
						Usage:       "Set a location",
						Description: "Manually set the GPS coordinates for the radio",
						Action:      setLocation,
						Flags: []cli.Flag{
							&cli.Float64Flag{
								Name:     "lat",
								Usage:    "Latitude",
								Required: true,
							},
							&cli.Float64Flag{
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
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "specify a port",
			},
		},
	}

	app.Run(os.Args)

}
