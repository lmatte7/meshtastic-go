package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/lmatte7/meshtastic-tui/internal/radio"
)

// Message types for async operations

type tickMsg time.Time

type connectSuccessMsg struct {
	Nodes    []radio.Node
	Channels []ChannelInfo
}

type connectErrorMsg string

type dataRefreshMsg struct {
	Nodes      []radio.Node
	Channels   []ChannelInfo
	ConfigData string
}

type newMessagesMsg []radio.Message

type messageSentMsg struct{}

type errorMsg string

// Commands

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func connectCmd(client *radio.Client, port string) tea.Cmd {
	return func() tea.Msg {
		err := client.Connect(port)
		if err != nil {
			return connectErrorMsg(err.Error())
		}

		// Load data from database after successful connection
		var nodes []radio.Node
		var channels []ChannelInfo

		// Get nodes from database
		dbNodes, err := client.GetNodesFromDB()
		if err == nil {
			// Convert db.Node to radio.Node
			for _, dbNode := range dbNodes {
				node := radio.Node{
					Num:          dbNode.NodeID,
					LongName:     dbNode.LongName,
					ShortName:    dbNode.ShortName,
					BatteryLevel: dbNode.BatteryLevel,
					Voltage:      dbNode.Voltage,
					Altitude:     dbNode.Altitude,
					Latitude:     dbNode.Latitude,
					Longitude:    dbNode.Longitude,
					ChannelUtil:  dbNode.ChannelUtil,
					AirUtilTx:    dbNode.AirUtilTx,
					LastHeard:    dbNode.LastHeard,
				}
				nodes = append(nodes, node)
			}
		}

		// Get channels from database (will be empty for now due to sync issues)
		dbChannels, err := client.GetChannelsFromDB()
		if err == nil {
			for _, dbChannel := range dbChannels {
				channel := ChannelInfo{
					Index: int32(dbChannel.Index),
					Name:  dbChannel.Name,
					Role:  dbChannel.Role,
				}
				channels = append(channels, channel)
			}
		}

		return connectSuccessMsg{
			Nodes:    nodes,
			Channels: channels,
		}
	}
}

func refreshDataCmd(client *radio.Client) tea.Cmd {
	return func() tea.Msg {
		nodes := []radio.Node{}
		channels := []ChannelInfo{}
		configData := ""

		// Get nodes
		if nodeData, err := client.GetRadioInfo(); err == nil && nodeData != nil {
			nodes = radio.ExtractNodes(nodeData)
		}

		// Get channels
		if channelData, err := client.GetChannels(); err == nil {
			channels = extractChannels(channelData)
		}

		// Get config
		if configs, moduleConfigs, err := client.GetRadioConfig(); err == nil {
			configData = formatConfig(configs, moduleConfigs)
		}

		return dataRefreshMsg{
			Nodes:      nodes,
			Channels:   channels,
			ConfigData: configData,
		}
	}
}

func checkMessagesCmd(client *radio.Client) tea.Cmd {
	return func() tea.Msg {
		responses, err := client.ReadResponse(false)
		if err != nil || len(responses) == 0 {
			return newMessagesMsg([]radio.Message{})
		}

		messages := radio.ExtractMessages(responses)
		return newMessagesMsg(messages)
	}
}

func sendMessageCmd(client *radio.Client, message string, to uint32, channel uint32) tea.Cmd {
	return func() tea.Msg {
		err := client.SendTextMessage(message, int64(to), int64(channel))
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to send: %s", err.Error()))
		}
		return messageSentMsg{}
	}
}

// Helper functions

func extractChannels(channelData []*gomeshproto.Channel) []ChannelInfo {
	channels := make([]ChannelInfo, 0, len(channelData))

	for _, ch := range channelData {
		if ch == nil {
			continue
		}

		settings := ch.GetSettings()
		if settings == nil {
			continue
		}

		role := "DISABLED"
		switch ch.GetRole() {
		case gomeshproto.Channel_PRIMARY:
			role = "PRIMARY"
		case gomeshproto.Channel_SECONDARY:
			role = "SECONDARY"
		}

		psk := ""
		if len(settings.Psk) > 0 {
			psk = fmt.Sprintf("%x", settings.Psk)
		}

		channels = append(channels, ChannelInfo{
			Index:    ch.GetIndex(),
			Name:     settings.GetName(),
			Role:     role,
			PSK:      psk,
			Uplink:   settings.GetUplinkEnabled(),
			Downlink: settings.GetDownlinkEnabled(),
		})
	}

	return channels
}

func formatConfig(configs []*gomeshproto.FromRadio_Config, moduleConfigs []*gomeshproto.FromRadio_ModuleConfig) string {
	var result string

	result += "=== Device Configuration ===\n\n"
	result += fmt.Sprintf("Received %d config items\n", len(configs))

	// Simplified config display - just show we got data
	for i, cfg := range configs {
		if cfg != nil {
			result += fmt.Sprintf("Config item %d: %T\n", i+1, cfg)
		}
	}

	result += "\n=== Module Configuration ===\n\n"
	result += fmt.Sprintf("Received %d module config items\n", len(moduleConfigs))

	for i, modCfg := range moduleConfigs {
		if modCfg != nil {
			result += fmt.Sprintf("Module config item %d: %T\n", i+1, modCfg)
		}
	}

	if len(configs) == 0 && len(moduleConfigs) == 0 {
		return "No configuration data available. Press 'r' to refresh."
	}

	return result
}
