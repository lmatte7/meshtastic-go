package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primaryColor   = lipgloss.Color("86")  // Cyan
	secondaryColor = lipgloss.Color("141") // Purple
	accentColor    = lipgloss.Color("120") // Green
	errorColor     = lipgloss.Color("196") // Red

	dimColor = lipgloss.Color("240") // Dim gray
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	viewStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	activeViewStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Background(lipgloss.Color("236")).
			Bold(true).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				Padding(0, 1)
)

func (m Model) renderHeader() string {
	title := titleStyle.Render("📡 Meshtastic TUI")

	status := ""
	if m.client.IsConnected() {
		status = lipgloss.NewStyle().Foreground(accentColor).Render("● Connected")
	} else {
		status = lipgloss.NewStyle().Foreground(errorColor).Render("○ Disconnected")
	}

	// Top line with title and status
	topLine := lipgloss.NewStyle().
		Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top,
			title,
			strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(status)-2)),
			status,
		))

	// View tabs
	views := []string{"Connect", "Messages (F1)", "Nodes (F2)", "Channels (F3)", "Config (F4)", "Help (F5)", "Logs (L)"}
	var tabs []string
	for i, v := range views {
		if View(i) == m.currentView {
			tabs = append(tabs, activeViewStyle.Render(" "+v+" "))
		} else {
			tabs = append(tabs, viewStyle.Render(" "+v+" "))
		}
	}

	tabBar := lipgloss.NewStyle().
		Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))

	separator := strings.Repeat("─", m.width)

	return lipgloss.JoinVertical(lipgloss.Left,
		topLine,
		tabBar,
		separator,
	)
}

func (m Model) renderFooter() string {
	help := ""
	switch m.currentView {
	case ViewConnect:
		help = "Enter: Connect | Ctrl+C/Q: Quit"
	case ViewMessages:
		if m.messagePaneFocus < 2 {
			help = "←→: Switch panes | ↑↓: Navigate list | Enter: Select | F1-F5: Switch view | R: Refresh | Q: Quit"
		} else {
			help = "Tab: Next field | Esc: Back to panes | Enter: Send | F1-F5: Switch view | R: Refresh | Q: Quit"
		}
	case ViewNodes, ViewChannels, ViewConfig:
		help = "↑↓: Scroll | F1-F5: Switch view | L: Logs | R: Refresh | Q: Quit"
	case ViewHelp:
		help = "F1-F5: Switch view | L: Logs | Q: Quit"
	case ViewLogs:
		help = "↑↓: Scroll logs | F1-F5: Switch view | R: Refresh logs | Q: Quit"
	}

	statusLine := ""
	if m.statusMsg != "" {
		statusLine = fmt.Sprintf("%s | %s", m.statusMsg, m.lastUpdate.Format("15:04:05"))
	}

	return strings.Repeat("─", m.width) + "\n" +
		statusStyle.Render(help) + "\n" +
		statusStyle.Render(statusLine)
}

func (m Model) renderConnect() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Connect to Meshtastic Device"))
	s.WriteString("\n\n")

	if m.connectError != "" {
		s.WriteString(errorStyle.Render("Error: " + m.connectError))
		s.WriteString("\n\n")
	}

	s.WriteString("Enter device port or IP address:\n\n")
	s.WriteString(inputStyle.Render(m.connectInput.View()))
	s.WriteString("\n\n")

	s.WriteString(lipgloss.NewStyle().Foreground(dimColor).Render("Examples:"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(dimColor).Render("  Serial: /dev/ttyUSB0, /dev/ttyACM0, COM3"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(dimColor).Render("  TCP/IP: 192.168.1.100"))
	s.WriteString("\n\n")

	if m.connecting {
		if m.syncProgress != "" {
			s.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render(m.syncProgress))
		} else {
			s.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render("Connecting..."))
		}
	}

	return s.String()
}

func (m Model) renderMessages() string {
	// Calculate pane widths (1/4, 1/4, 1/2)
	totalWidth := m.width - 6 // Account for borders and padding
	channelPaneWidth := totalWidth / 4
	nodePaneWidth := totalWidth / 4
	messagePaneWidth := totalWidth - channelPaneWidth - nodePaneWidth

	// Channel pane
	channelPane := m.renderChannelPane(channelPaneWidth)

	// Node pane
	nodePane := m.renderNodePane(nodePaneWidth)

	// Message pane
	messagePane := m.renderMessagePane(messagePaneWidth)

	// Combine panes horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top,
		channelPane,
		" ", // Separator
		nodePane,
		" ", // Separator
		messagePane,
	)
}

func (m Model) renderChannelPane(width int) string {
	// Pane style
	paneStyle := lipgloss.NewStyle().
		Width(width).
		Height(m.height - 6). // Account for header/footer
		Border(lipgloss.RoundedBorder()).
		Padding(1)

	if m.messagePaneFocus == 0 {
		paneStyle = paneStyle.BorderForeground(accentColor)
	} else {
		paneStyle = paneStyle.BorderForeground(dimColor)
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Channels"))
	content.WriteString("\n\n")

	if len(m.channels) == 0 {
		content.WriteString("No channels")
	} else {
		for i, channel := range m.channels {
			style := lipgloss.NewStyle()
			if i == m.selectedChannelIdx && m.messagePaneFocus == 0 {
				style = style.Background(accentColor).Foreground(lipgloss.Color("0"))
			}

			name := sanitizeUTF8(channel.Name)
			if name == "" {
				name = fmt.Sprintf("Channel %d", channel.Index)
			}
			if len(name) > width-6 {
				name = name[:width-9] + "..."
			}

			content.WriteString(style.Render(fmt.Sprintf("%d: %s", channel.Index, name)))
			content.WriteString("\n")
		}
	}

	return paneStyle.Render(content.String())
}

func (m Model) renderNodePane(width int) string {
	// Pane style
	paneStyle := lipgloss.NewStyle().
		Width(width).
		Height(m.height - 6). // Account for header/footer
		Border(lipgloss.RoundedBorder()).
		Padding(1)

	if m.messagePaneFocus == 1 {
		paneStyle = paneStyle.BorderForeground(accentColor)
	} else {
		paneStyle = paneStyle.BorderForeground(dimColor)
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Nodes"))
	content.WriteString("\n\n")

	if len(m.nodes) == 0 {
		content.WriteString("No nodes")
	} else {
		for i, node := range m.nodes {
			style := lipgloss.NewStyle()
			if i == m.selectedNodeIdx && m.messagePaneFocus == 1 {
				style = style.Background(accentColor).Foreground(lipgloss.Color("0"))
			}

			name := sanitizeUTF8(node.LongName)
			if name == "" {
				name = sanitizeUTF8(node.ShortName)
			}
			if name == "" {
				name = "Unknown"
			}
			if len(name) > width-12 {
				name = name[:width-15] + "..."
			}

			content.WriteString(style.Render(fmt.Sprintf("%d: %s", node.Num, name)))
			content.WriteString("\n")
		}
	}

	return paneStyle.Render(content.String())
}

func (m Model) renderMessagePane(width int) string {
	// Pane style
	paneStyle := lipgloss.NewStyle().
		Width(width).
		Height(m.height - 6). // Account for header/footer
		Border(lipgloss.RoundedBorder()).
		Padding(1)

	if m.messagePaneFocus == 2 {
		paneStyle = paneStyle.BorderForeground(accentColor)
	} else {
		paneStyle = paneStyle.BorderForeground(dimColor)
	}

	// Header
	header := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Messages (%d)", len(m.messages)))

	// Message viewport (adjust height for inputs)
	viewportHeight := m.height - 12 // Account for header, footer, inputs
	messageViewport := lipgloss.NewStyle().
		Width(width - 4).
		Height(viewportHeight).
		Render(m.messageViewport.View())

	// Input fields
	msgStyle := inputStyle
	if m.messageInputFocus == 0 && m.messagePaneFocus == 2 {
		msgStyle = focusedInputStyle
	}
	messageField := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(8).Render("Msg:"),
		msgStyle.Width(width-12).Render(m.messageInput.View()),
	)

	toStyle := inputStyle
	if m.messageInputFocus == 1 && m.messagePaneFocus == 2 {
		toStyle = focusedInputStyle
	}
	toField := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(8).Render("To:"),
		toStyle.Width(width-12).Render(m.messageToInput.View()),
	)

	chanStyle := inputStyle
	if m.messageInputFocus == 2 && m.messagePaneFocus == 2 {
		chanStyle = focusedInputStyle
	}
	channelField := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(8).Render("Chan:"),
		chanStyle.Width(width-12).Render(m.messageChanInput.View()),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		messageViewport,
		"",
		messageField,
		toField,
		channelField,
	)

	return paneStyle.Render(content)
}

func (m Model) renderNodes() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Nodes (%d)", len(m.nodes))))
	s.WriteString("\n\n")

	if len(m.nodes) == 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(dimColor).Render("No nodes found. Press 'r' to refresh."))
	} else {
		s.WriteString(m.nodeViewport.View())
	}

	return s.String()
}

func (m Model) renderChannels() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Channels (%d)", len(m.channels))))
	s.WriteString("\n\n")

	if len(m.channels) == 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(dimColor).Render("No channels found. Press 'r' to refresh."))
	} else {
		s.WriteString(m.channelViewport.View())
	}

	return s.String()
}

func (m Model) renderConfig() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Configuration"))
	s.WriteString("\n\n")

	if m.configData == "" {
		s.WriteString(lipgloss.NewStyle().Foreground(dimColor).Render("No configuration data. Press 'r' to refresh."))
	} else {
		s.WriteString(m.configViewport.View())
	}

	return s.String()
}

func (m Model) renderHelp() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Help & Keyboard Shortcuts"))
	s.WriteString("\n\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Navigation"))
	s.WriteString("\n")
	s.WriteString("  F1          - Messages view\n")
	s.WriteString("  F2          - Nodes view\n")
	s.WriteString("  F3          - Channels view\n")
	s.WriteString("  F4          - Configuration view\n")
	s.WriteString("  F5          - Help (this screen)\n")
	s.WriteString("  R           - Refresh data from device (when not typing in input field)\n")
	s.WriteString("  Q / Ctrl+C  - Quit application\n")
	s.WriteString("\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Messages View"))
	s.WriteString("\n")
	s.WriteString("  ← / →       - Switch between Channel/Node/Input panes\n")
	s.WriteString("  ↑ / ↓       - Navigate channels/nodes in left panes\n")
	s.WriteString("  Enter       - Select channel/node (auto-fills inputs)\n")
	s.WriteString("  Tab         - Move to next input field (in input pane)\n")
	s.WriteString("  Shift+Tab   - Move to previous input field\n")
	s.WriteString("  Escape      - Go back to channel/node panes\n")
	s.WriteString("  Enter       - Send message (from input pane)\n")
	s.WriteString("\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Message Fields"))
	s.WriteString("\n")
	s.WriteString("  Message     - Text to send (max 237 chars)\n")
	s.WriteString("  To          - Node ID (0 or empty = broadcast to all)\n")
	s.WriteString("  Channel     - Channel number (0 or empty = default)\n")
	s.WriteString("\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Scrollable Views"))
	s.WriteString("\n")
	s.WriteString("  ↑ / ↓       - Scroll up/down\n")
	s.WriteString("  PgUp / PgDn - Scroll page up/down\n")
	s.WriteString("  Home / End  - Jump to top/bottom\n")
	s.WriteString("\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Features"))
	s.WriteString("\n")
	s.WriteString("  • Real-time message monitoring\n")
	s.WriteString("  • Send messages to specific nodes or broadcast\n")
	s.WriteString("  • View all nodes with battery and location info\n")
	s.WriteString("  • View channel configurations\n")
	s.WriteString("  • View device configuration\n")
	s.WriteString("  • UTF-8 safe string handling\n")
	s.WriteString("  • Scrollable content in all views\n")
	s.WriteString("\n")

	return s.String()
}

func (m Model) renderLogs() string {
	return m.logViewport.View()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
