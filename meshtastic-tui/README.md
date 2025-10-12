# Meshtastic TUI

A beautiful, comprehensive, and interactive Terminal User Interface (TUI) for Meshtastic devices.

![Meshtastic TUI](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

✨ **Beautiful Interface** - Modern TUI built with Bubble Tea and Lipgloss
📨 **Real-time Messaging** - View and send messages on the mesh network
🌐 **Node Monitoring** - See all nodes with battery, position, and metrics
📡 **Channel Management** - View channel configurations and settings
⚙️ **Configuration View** - Check radio settings and configuration
⌨️ **Keyboard Navigation** - Full keyboard control with intuitive shortcuts
🎨 **Color-coded Display** - Easy-to-read color scheme for better UX

## Installation

### Prerequisites

- Go 1.21 or higher
- A Meshtastic device connected via serial or accessible via TCP/IP
- ESP32 USB drivers (if using serial connection)

### Build from Source

```bash
cd meshtastic-tui
go mod download
go build -o meshtastic-tui
```

### Run

```bash
./meshtastic-tui
```

## Usage

### Connecting to Your Device

1. Launch the application
2. Enter your device port or IP address:
   - **Serial**: `/dev/ttyUSB0`, `/dev/cu.SLAB_USBtoUART`, `COM3`
   - **TCP/IP**: `192.168.1.100`, `meshtastic.local`
3. Press `Enter` to connect

### Navigation

The TUI has multiple tabs for different functionality:

#### 🔌 Connect Tab
- Enter device connection details
- View connection status
- See connection examples

#### 💬 Messages Tab (F1)
- View real-time messages from the mesh
- Send broadcast messages to all nodes
- See message history with timestamps
- Auto-refreshes every second

#### 🌐 Nodes Tab (F2)
- View all nodes on the mesh network
- See node details:
  - Node ID and name
  - Battery level and voltage
  - GPS coordinates (latitude/longitude)
  - Device metrics

#### 📡 Channels Tab (F3)
- View all configured channels
- See channel settings:
  - Channel index and name
  - Role (PRIMARY, SECONDARY, etc.)
  - Uplink/Downlink status
  - PSK (Pre-Shared Key)

#### ⚙️ Config Tab (F4)
- View radio configuration
- Device settings
- Position/GPS settings
- LoRa configuration
- Module settings

#### ❓ Help Tab (F5)
- Keyboard shortcuts reference
- Feature overview
- About information

## Keyboard Shortcuts

### Navigation
- `Tab` / `Shift+Tab` - Switch between tabs
- `F1` - Go to Messages
- `F2` - Go to Nodes
- `F3` - Go to Channels
- `F4` - Go to Config
- `F5` - Go to Help

### Actions
- `R` - Refresh data from radio
- `Enter` - Connect to device / Send message
- `Q` / `Ctrl+C` - Quit application

## Architecture

This application is built as a **standalone Go application** that imports the `meshtastic-go` library as a dependency. It does not modify the original meshtastic-go codebase.

### Project Structure

```
meshtastic-tui/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── internal/
│   ├── radio/
│   │   └── client.go      # Radio client wrapper
│   └── ui/
│       ├── model.go       # Main TUI model
│       ├── commands.go    # Async commands
│       └── views.go       # View rendering
└── README.md
```

### Components

#### Radio Client (`internal/radio/client.go`)
- Wraps the `gomesh.Radio` type
- Provides connection management
- Helper functions for data extraction
- Message and node data structures

#### UI Model (`internal/ui/model.go`)
- Main Bubble Tea model
- State management
- Tab navigation
- Input handling

#### Commands (`internal/ui/commands.go`)
- Async operations (connect, refresh, send)
- Message types for state updates
- Data formatting helpers

#### Views (`internal/ui/views.go`)
- Rendering functions for each tab
- Styling with Lipgloss
- Color scheme and layout

## Features in Detail

### Real-time Message Monitoring
The Messages tab automatically checks for new messages every second when connected. Messages are displayed with:
- Timestamp
- Sender node ID
- Recipient (or "ALL" for broadcast)
- Message content

### Node Information
The Nodes tab displays comprehensive information about all nodes on the mesh:
- Node identification (ID and name)
- Power status (battery percentage and voltage)
- Location data (GPS coordinates)
- Network metrics (channel utilization, air time)

### Channel Configuration
View all configured channels with their settings:
- Channel index and role
- Uplink/Downlink capabilities
- Pre-Shared Keys (PSK) for encryption
- Channel names

### Radio Configuration
Inspect your radio's configuration including:
- Device role and settings
- GPS/Position configuration
- LoRa parameters (region, modem preset, hop limit)
- Module settings (MQTT, telemetry, etc.)

## Dependencies

This application uses the following libraries:

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Styling and layout
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - TUI components
- **[meshtastic-go](https://github.com/lmatte7/meshtastic-go)** - Meshtastic library
- **[gomesh](https://github.com/lmatte7/gomesh)** - Mesh protocol implementation

## Troubleshooting

### Connection Issues

**Serial Port Not Found**
- Ensure ESP32 USB drivers are installed
- Check that the device is properly connected
- Verify the port name (use `ls /dev/tty*` on Linux/Mac or Device Manager on Windows)

**TCP Connection Failed**
- Verify the device is on the same network
- Check the IP address is correct
- Ensure the device has WiFi enabled

### Display Issues

**Colors Not Showing**
- Ensure your terminal supports 256 colors
- Try a modern terminal emulator (iTerm2, Windows Terminal, etc.)

**Layout Problems**
- Resize your terminal window (minimum 80x24 recommended)
- The TUI adapts to window size changes automatically

## Contributing

This is a standalone application built on top of the meshtastic-go library. Contributions are welcome!

## License

This project follows the same license as the meshtastic-go library it depends on.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm
- Uses [meshtastic-go](https://github.com/lmatte7/meshtastic-go) by Lucas Matte
- Inspired by the Meshtastic community

## Support

For issues related to:
- **TUI application**: Open an issue in this repository
- **Meshtastic library**: See [meshtastic-go issues](https://github.com/lmatte7/meshtastic-go/issues)
- **Meshtastic devices**: See [Meshtastic documentation](https://meshtastic.org)

