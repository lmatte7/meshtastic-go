package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmatte7/meshtastic-tui/internal/radio"
)

// View represents different views
type View int

const (
	ViewConnect View = iota
	ViewMessages
	ViewNodes
	ViewChannels
	ViewConfig
	ViewHelp
	ViewLogs
)

// Model is the main application model
type Model struct {
	client      *radio.Client
	currentView View
	width       int
	height      int

	// Connection
	connectInput textinput.Model
	connecting   bool
	connectError string
	autoConnect  string // Auto-connect device path
	syncProgress string // Current sync progress message

	// Messages
	messages          []radio.Message
	messageViewport   viewport.Model
	messageInput      textinput.Model
	messageToInput    textinput.Model
	messageChanInput  textinput.Model
	messageInputFocus int // 0=message, 1=to, 2=channel

	// Message view panes
	selectedChannelIdx int // Selected channel for messaging
	selectedNodeIdx    int // Selected node for DM
	messagePaneFocus   int // 0=channels, 1=nodes, 2=inputs

	// Nodes
	nodes        []radio.Node
	nodeViewport viewport.Model
	selectedNode int

	// Channels
	channels        []ChannelInfo
	channelViewport viewport.Model
	selectedChannel int

	// Config
	configData     string
	configViewport viewport.Model

	// Status
	statusMsg  string
	lastUpdate time.Time

	// Logs
	logViewport viewport.Model
	logContent  string
}

type ChannelInfo struct {
	Index    int32
	Name     string
	Role     string
	PSK      string
	Uplink   bool
	Downlink bool
}

// Ensure string is valid UTF-8
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "�")
}

// NewModel creates a new model
func NewModel() (Model, error) {
	client, err := radio.NewClient()
	if err != nil {
		return Model{}, fmt.Errorf("failed to create radio client: %w", err)
	}

	connectInput := textinput.New()
	connectInput.Placeholder = "/dev/ttyUSB0 or 192.168.1.100"
	connectInput.Focus()
	connectInput.Width = 50

	messageInput := textinput.New()
	messageInput.Placeholder = "Message text..."
	messageInput.Width = 60
	messageInput.Focus()

	messageToInput := textinput.New()
	messageToInput.Placeholder = "To (node ID, 0=all)"
	messageToInput.Width = 20

	messageChanInput := textinput.New()
	messageChanInput.Placeholder = "Channel (0=default)"
	messageChanInput.Width = 20

	return Model{
		client:            client,
		currentView:       ViewConnect,
		connectInput:      connectInput,
		messageInput:      messageInput,
		messageToInput:    messageToInput,
		messageChanInput:  messageChanInput,
		messageInputFocus: 0,
		messages:          []radio.Message{},
		nodes:             []radio.Node{},
		channels:          []ChannelInfo{},
		messageViewport:   viewport.New(0, 0),
		nodeViewport:      viewport.New(0, 0),
		channelViewport:   viewport.New(0, 0),
		configViewport:    viewport.New(0, 0),
		logViewport:       viewport.New(0, 0),
		selectedNode:      -1,
		selectedChannel:   -1,
		lastUpdate:        time.Now(),
	}, nil
}

// SetAutoConnect sets the device path for auto-connection
func (m *Model) SetAutoConnect(devicePath string) {
	m.autoConnect = devicePath
}

// loadLogContent loads the log file content
func (m *Model) loadLogContent() {
	content, err := os.ReadFile("./tmp.log")
	if err != nil {
		m.logContent = fmt.Sprintf("Error reading log file: %v", err)
	} else {
		m.logContent = string(content)
	}
	m.logViewport.SetContent(m.logContent)
}

func (m Model) Init() tea.Cmd {
	// If auto-connect is set, start connection immediately
	if m.autoConnect != "" {
		return tea.Batch(
			textinput.Blink,
			tickCmd(),
			connectCmd(m.client, m.autoConnect),
		)
	}
	return tea.Batch(textinput.Blink, tickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		vpHeight := m.height - 10
		m.messageViewport.Width = m.width - 4
		m.messageViewport.Height = vpHeight
		m.nodeViewport.Width = m.width - 4
		m.nodeViewport.Height = vpHeight
		m.channelViewport.Width = m.width - 4
		m.channelViewport.Height = vpHeight
		m.configViewport.Width = m.width - 4
		m.configViewport.Height = vpHeight
		m.logViewport.Width = m.width - 4
		m.logViewport.Height = vpHeight

		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		// Always-available keys
		switch msg.String() {
		case "ctrl+c":
			m.client.Disconnect()
			return m, tea.Quit
		}

		// Check if we're typing in an input field
		inInputField := false
		if m.currentView == ViewConnect && m.connectInput.Focused() {
			inInputField = true
		} else if m.currentView == ViewMessages && (m.messageInput.Focused() || m.messageToInput.Focused() || m.messageChanInput.Focused()) {
			inInputField = true
		}

		// Global keys (only when not typing)
		if !inInputField {
			switch msg.String() {
			case "q":
				m.client.Disconnect()
				return m, tea.Quit
			case "r":
				if m.client.IsConnected() {
					return m, refreshDataCmd(m.client)
				}
				return m, nil
			}
		}

		// Navigation keys (always available)
		switch msg.String() {
		case "f1":
			if m.client.IsConnected() {
				m.currentView = ViewMessages
				m.messagePaneFocus = 0 // Start with channels pane
				m.messageInput.Blur()
				m.messageToInput.Blur()
				m.messageChanInput.Blur()
			}
			return m, nil
		case "f2":
			if m.client.IsConnected() {
				m.currentView = ViewNodes
				m.updateViewportContent()
			}
			return m, nil
		case "f3":
			if m.client.IsConnected() {
				m.currentView = ViewChannels
				m.updateViewportContent()
			}
			return m, nil
		case "f4":
			if m.client.IsConnected() {
				m.currentView = ViewConfig
			}
			return m, nil
		case "f5":
			m.currentView = ViewHelp
			return m, nil
		case "f6", "l":
			m.currentView = ViewLogs
			m.loadLogContent()
			return m, nil
		}

		// View-specific keys
		switch m.currentView {
		case ViewConnect:
			switch msg.String() {
			case "enter":
				if !m.connecting && m.connectInput.Value() != "" {
					m.connecting = true
					m.connectError = ""
					return m, connectCmd(m.client, m.connectInput.Value())
				}
			}
			m.connectInput, cmd = m.connectInput.Update(msg)
			return m, cmd

		case ViewMessages:
			// Handle pane navigation first (when not in input mode)
			if m.messagePaneFocus < 2 {
				switch msg.String() {
				case "escape":
					// Reset to channel pane
					m.messagePaneFocus = 0
					return m, nil
				case "left":
					if m.messagePaneFocus > 0 {
						m.messagePaneFocus--
					}
					return m, nil
				case "right":
					if m.messagePaneFocus < 1 {
						m.messagePaneFocus++
					}
					return m, nil
				case "up":
					if m.messagePaneFocus == 0 && len(m.channels) > 0 {
						// Navigate channels
						if m.selectedChannelIdx > 0 {
							m.selectedChannelIdx--
						}
					} else if m.messagePaneFocus == 1 && len(m.nodes) > 0 {
						// Navigate nodes
						if m.selectedNodeIdx > 0 {
							m.selectedNodeIdx--
						}
					}
					return m, nil
				case "down":
					if m.messagePaneFocus == 0 && len(m.channels) > 0 {
						// Navigate channels
						if m.selectedChannelIdx < len(m.channels)-1 {
							m.selectedChannelIdx++
						}
					} else if m.messagePaneFocus == 1 && len(m.nodes) > 0 {
						// Navigate nodes
						if m.selectedNodeIdx < len(m.nodes)-1 {
							m.selectedNodeIdx++
						}
					}
					return m, nil
				case "enter":
					if m.messagePaneFocus == 0 && len(m.channels) > 0 {
						// Select channel and auto-fill channel input
						selectedChannel := m.channels[m.selectedChannelIdx]
						m.messageChanInput.SetValue(fmt.Sprintf("%d", selectedChannel.Index))
						m.messagePaneFocus = 2 // Move to inputs
						m.messageInputFocus = 0
						m.messageInput.Focus()
						return m, nil
					} else if m.messagePaneFocus == 1 && len(m.nodes) > 0 {
						// Select node and auto-fill to input
						selectedNode := m.nodes[m.selectedNodeIdx]
						m.messageToInput.SetValue(fmt.Sprintf("%d", selectedNode.Num))
						m.messagePaneFocus = 2 // Move to inputs
						m.messageInputFocus = 0
						m.messageInput.Focus()
						return m, nil
					}
					return m, nil
				}
			}

			// Handle input field navigation and actions
			switch msg.String() {
			case "escape":
				// Blur all inputs and go back to channel pane
				m.messageInput.Blur()
				m.messageToInput.Blur()
				m.messageChanInput.Blur()
				m.messagePaneFocus = 0
				return m, nil

			case "tab":
				m.messageInputFocus = (m.messageInputFocus + 1) % 3
				m.messageInput.Blur()
				m.messageToInput.Blur()
				m.messageChanInput.Blur()
				switch m.messageInputFocus {
				case 0:
					m.messageInput.Focus()
				case 1:
					m.messageToInput.Focus()
				case 2:
					m.messageChanInput.Focus()
				}
				return m, nil
			case "shift+tab":
				m.messageInputFocus = (m.messageInputFocus + 2) % 3
				m.messageInput.Blur()
				m.messageToInput.Blur()
				m.messageChanInput.Blur()
				switch m.messageInputFocus {
				case 0:
					m.messageInput.Focus()
				case 1:
					m.messageToInput.Focus()
				case 2:
					m.messageChanInput.Focus()
				}
				return m, nil
			case "enter":
				if m.messageInput.Value() != "" {
					to := uint32(0)
					if m.messageToInput.Value() != "" {
						if val, err := strconv.ParseUint(m.messageToInput.Value(), 10, 32); err == nil {
							to = uint32(val)
						}
					}
					channel := uint32(0)
					if m.messageChanInput.Value() != "" {
						if val, err := strconv.ParseUint(m.messageChanInput.Value(), 10, 32); err == nil {
							channel = uint32(val)
						}
					}
					return m, sendMessageCmd(m.client, m.messageInput.Value(), to, channel)
				}
				return m, nil
			case "up", "down", "pgup", "pgdown":
				m.messageViewport, cmd = m.messageViewport.Update(msg)
				return m, cmd
			}

			switch m.messageInputFocus {
			case 0:
				m.messageInput, cmd = m.messageInput.Update(msg)
			case 1:
				m.messageToInput, cmd = m.messageToInput.Update(msg)
			case 2:
				m.messageChanInput, cmd = m.messageChanInput.Update(msg)
			}
			return m, cmd

		case ViewNodes:
			m.nodeViewport, cmd = m.nodeViewport.Update(msg)
			return m, cmd

		case ViewChannels:
			m.channelViewport, cmd = m.channelViewport.Update(msg)
			return m, cmd

		case ViewConfig:
			m.configViewport, cmd = m.configViewport.Update(msg)
			return m, cmd

		case ViewLogs:
			// Handle refresh for logs
			if msg.String() == "r" {
				m.loadLogContent()
				return m, nil
			}
			m.logViewport, cmd = m.logViewport.Update(msg)
			return m, cmd
		}

	case tickMsg:
		if m.client.IsConnected() {
			// Always check for new messages regardless of current view
			cmds = append(cmds, checkMessagesCmd(m.client))
		}
		cmds = append(cmds, tickCmd())
		return m, tea.Batch(cmds...)

	case connectSuccessMsg:
		m.connecting = false
		m.connectError = ""
		m.statusMsg = "Connected!"
		m.lastUpdate = time.Now()
		m.currentView = ViewMessages
		m.messageInput.Focus()

		// Load data from the connect success message
		m.nodes = msg.Nodes
		m.channels = msg.Channels
		m.updateViewportContent()

		return m, tea.Batch(checkMessagesCmd(m.client), tickCmd())

	case connectErrorMsg:
		m.connecting = false
		m.connectError = string(msg)
		m.statusMsg = "Connection failed"
		m.syncProgress = ""
		return m, nil

	case syncProgressMsg:
		m.syncProgress = msg.Message
		if msg.Error != "" {
			m.syncProgress = fmt.Sprintf("⚠️ %s: %s", msg.Stage, msg.Error)
		}
		return m, nil

	case dataRefreshMsg:
		m.nodes = msg.Nodes
		m.channels = msg.Channels
		m.configData = msg.ConfigData
		m.configViewport.SetContent(m.configData)
		m.updateViewportContent()
		m.statusMsg = "Data refreshed"
		m.lastUpdate = time.Now()
		return m, nil

	case newMessagesMsg:
		if len(msg) > 0 {
			m.messages = append(m.messages, msg...)
			m.updateViewportContent()
			m.statusMsg = fmt.Sprintf("Received %d message(s)", len(msg))
			m.lastUpdate = time.Now()
		}
		return m, nil

	case messageSentMsg:
		m.messageInput.SetValue("")
		m.messageToInput.SetValue("")
		m.messageChanInput.SetValue("")
		m.statusMsg = "Message sent!"
		m.lastUpdate = time.Now()
		return m, nil

	case errorMsg:
		m.statusMsg = string(msg)
		m.lastUpdate = time.Now()
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	var content string
	switch m.currentView {
	case ViewConnect:
		content = m.renderConnect()
	case ViewMessages:
		content = m.renderMessages()
	case ViewNodes:
		content = m.renderNodes()
	case ViewChannels:
		content = m.renderChannels()
	case ViewConfig:
		content = m.renderConfig()
	case ViewHelp:
		content = m.renderHelp()
	case ViewLogs:
		content = m.renderLogs()
	}

	// Calculate available height for content
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := m.height - headerHeight - footerHeight - 1

	// Style the content area
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		contentStyle.Render(content),
		footer,
	)
}

// updateViewportContent updates viewport content based on current data
func (m *Model) updateViewportContent() {
	switch m.currentView {
	case ViewMessages:
		var content strings.Builder
		for _, msg := range m.messages {
			timestamp := msg.Timestamp.Format("15:04:05")
			fromStr := fmt.Sprintf("%d", msg.From)
			toStr := "ALL"
			if msg.To != 0xFFFFFFFF {
				toStr = fmt.Sprintf("%d", msg.To)
			}
			payload := sanitizeUTF8(msg.Payload)
			content.WriteString(fmt.Sprintf("[%s] %s → %s: %s\n", timestamp, fromStr, toStr, payload))
		}
		m.messageViewport.SetContent(content.String())

	case ViewNodes:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("%-12s %-20s %-10s %-10s %-12s %-12s\n",
			"Node ID", "Name", "Battery", "Voltage", "Latitude", "Longitude"))
		content.WriteString(strings.Repeat("─", 80) + "\n")

		for _, node := range m.nodes {
			name := sanitizeUTF8(node.LongName)
			if name == "" {
				name = sanitizeUTF8(node.ShortName)
			}
			if name == "" {
				name = "Unknown"
			}
			if len(name) > 20 {
				name = name[:17] + "..."
			}

			battery := "N/A"
			if node.BatteryLevel > 0 {
				battery = fmt.Sprintf("%d%%", node.BatteryLevel)
			}

			voltage := "N/A"
			if node.Voltage > 0 {
				voltage = fmt.Sprintf("%.2fV", node.Voltage)
			}

			lat := "N/A"
			if node.Latitude != 0 {
				lat = fmt.Sprintf("%.4f", float64(node.Latitude)/1e7)
			}

			lon := "N/A"
			if node.Longitude != 0 {
				lon = fmt.Sprintf("%.4f", float64(node.Longitude)/1e7)
			}

			content.WriteString(fmt.Sprintf("%-12d %-20s %-10s %-10s %-12s %-12s\n",
				node.Num, name, battery, voltage, lat, lon))
		}
		m.nodeViewport.SetContent(content.String())

	case ViewChannels:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("%-8s %-20s %-15s %-10s %-10s\n",
			"Index", "Name", "Role", "Uplink", "Downlink"))
		content.WriteString(strings.Repeat("─", 70) + "\n")

		for _, ch := range m.channels {
			name := sanitizeUTF8(ch.Name)
			if name == "" {
				name = "Default"
			}
			if len(name) > 20 {
				name = name[:17] + "..."
			}

			uplink := "No"
			if ch.Uplink {
				uplink = "Yes"
			}
			downlink := "No"
			if ch.Downlink {
				downlink = "Yes"
			}

			content.WriteString(fmt.Sprintf("%-8d %-20s %-15s %-10s %-10s\n",
				ch.Index, name, ch.Role, uplink, downlink))

			if ch.PSK != "" {
				content.WriteString(fmt.Sprintf("  PSK: %s\n", ch.PSK))
			}
		}
		m.channelViewport.SetContent(content.String())
	}
}
