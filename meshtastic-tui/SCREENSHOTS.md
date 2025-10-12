# Meshtastic TUI Screenshots & Examples

This document provides visual examples and descriptions of the TUI interface.

## Overview

The Meshtastic TUI provides a full-screen terminal interface with multiple tabs for different functionality.

## Interface Layout

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 📡 Meshtastic TUI                          ● Connected: /dev/ttyUSB0    │
│ Connect | Messages (F1) | Nodes (F2) | Channels (F3) | Config (F4) | Help (F5) │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│                         [TAB CONTENT AREA]                              │
│                                                                         │
│                                                                         │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ Message sent | Last update: 14:23:45      Tab: Switch | R: Refresh | Q: Quit │
└─────────────────────────────────────────────────────────────────────────┘
```

## Tab Views

### 1. Connect Tab

**Purpose**: Connect to your Meshtastic device

```
Connect to Meshtastic Device
═════════════════════════════════════════════════════════════════════════

Enter the port or IP address of your Meshtastic device:

> /dev/ttyUSB0_

Examples:
  • Serial: /dev/ttyUSB0, /dev/cu.SLAB_USBtoUART, COM3
  • TCP: 192.168.1.100, meshtastic.local

Press Enter to connect
```

**Features**:
- Text input for port/IP address
- Examples for different platforms
- Connection status and error messages
- Auto-switches to Messages tab on success

### 2. Messages Tab (F1)

**Purpose**: View and send messages in real-time

```
Messages
═════════════════════════════════════════════════════════════════════════

[14:23:12] 123456789 → ALL: Hello mesh!
[14:23:45] 987654321 → ALL: Hi there!
[14:24:01] 123456789 → 555555555: Direct message
[14:24:15] 555555555 → 123456789: Reply

─────────────────────────────────────────────────────────────────────────

Send message (broadcast to all): Testing 123_
Press Enter to send
```

**Features**:
- Real-time message display with timestamps
- Shows sender and recipient (or "ALL" for broadcast)
- Last 10 messages shown
- Input field for sending messages
- Auto-refreshes every second

### 3. Nodes Tab (F2)

**Purpose**: View all nodes on the mesh network

```
Mesh Nodes (5)
═════════════════════════════════════════════════════════════════════════

Node ID      Name                 Battery    Voltage    Latitude     Longitude
─────────────────────────────────────────────────────────────────────────
123456789    Base Station         95%        4.15V      37.7749      -122.4194
987654321    Mobile Node 1        78%        3.92V      37.7750      -122.4195
555555555    Repeater             100%       4.20V      37.7751      -122.4196
111111111    Portable             45%        3.65V      N/A          N/A
222222222    Remote               N/A        N/A        37.7752      -122.4197
```

**Features**:
- Table view of all nodes
- Node ID and name
- Battery level and voltage
- GPS coordinates (if available)
- Shows "N/A" for missing data

### 4. Channels Tab (F3)

**Purpose**: View channel configurations

```
Channels (3)
═════════════════════════════════════════════════════════════════════════

Index    Name                 Role            Uplink     Downlink
─────────────────────────────────────────────────────────────────────────
0        Default              PRIMARY         Yes        Yes
  PSK: 1234567890abcdef1234567890abcdef

1        Secondary            SECONDARY       Yes        Yes
  PSK: fedcba0987654321fedcba0987654321

2        Admin                SECONDARY       No         Yes
  PSK: abcdef1234567890abcdef1234567890
```

**Features**:
- Channel index and name
- Role (PRIMARY, SECONDARY, etc.)
- Uplink/Downlink status
- PSK (encryption key) display
- Formatted table layout

### 5. Config Tab (F4)

**Purpose**: View radio configuration

```
Radio Configuration
═════════════════════════════════════════════════════════════════════════

Device Config:
  Role: CLIENT
  Serial Enabled: true
  Rebroadcast Mode: ALL_SKIP_DECODING

Position Config:
  GPS Enabled: true
  GPS Update Interval: 120
  Position Broadcast Secs: 900

LoRa Config:
  Region: US
  Modem Preset: LONG_FAST
  Hop Limit: 3
  Tx Enabled: true

Module Settings:
  MQTT Enabled: false
  Telemetry Update Interval: 900
```

**Features**:
- Device settings
- Position/GPS configuration
- LoRa parameters
- Module settings
- Organized by category

### 6. Help Tab (F5)

**Purpose**: Show keyboard shortcuts and help

```
Help & Keyboard Shortcuts
═════════════════════════════════════════════════════════════════════════

Navigation
────────────────────────────────────────────────────────────────
  Tab / Shift+Tab      Switch between tabs
  F1                   Go to Messages
  F2                   Go to Nodes
  F3                   Go to Channels
  F4                   Go to Config
  F5                   Go to Help

Actions
────────────────────────────────────────────────────────────────
  R                    Refresh data from radio
  Enter                Connect / Send message
  Q / Ctrl+C           Quit application

About
────────────────────────────────────────────────────────────────
Meshtastic TUI - A beautiful terminal interface for Meshtastic devices

This application provides an interactive way to:
  • View and send messages on the mesh network
  • Monitor nodes and their status
  • View channel configurations
  • Check radio settings

Built with Bubble Tea and the meshtastic-go library
```

**Features**:
- Complete keyboard shortcut reference
- Organized by category
- About information
- Always accessible (F5)

## Color Scheme

The TUI uses a carefully chosen color palette:

- **Primary** (Cyan): Titles, important text, node IDs
- **Secondary** (Purple): Active tabs, recipient IDs
- **Accent** (Green): Success messages, connected status
- **Error** (Red): Error messages, disconnected status
- **Warning** (Orange): Warning messages, connecting status
- **Text** (Light Gray): Normal text
- **Dim** (Gray): Secondary text, help text
- **Background** (Dark Gray): Backgrounds
- **Border** (Medium Gray): Borders and dividers

## Status Bar

The status bar at the bottom shows:

**Left side**:
- Last action performed ("Message sent", "Data refreshed", etc.)
- Last update timestamp

**Right side**:
- Quick help: "Tab: Switch | R: Refresh | Q: Quit"

## Interactive Elements

### Text Input Fields
- Cursor visible and blinking
- Character limit shown (e.g., 237 chars for messages)
- Placeholder text when empty
- Styled with background color

### Tables
- Header row in bold with primary color
- Alternating row styles (could be added)
- Aligned columns
- Divider lines between sections

### Tabs
- Active tab highlighted with background color
- Inactive tabs shown in dim color
- Function key shortcuts shown (F1-F5)
- Easy visual identification

## Responsive Design

The TUI adapts to terminal size:
- Minimum recommended: 80x24
- Adjusts content area based on window size
- Tables truncate long text with "..."
- Scrolling for long content (future enhancement)

## Real-Time Updates

### Auto-Refresh
- Messages tab refreshes every second
- Shows new messages as they arrive
- Non-blocking updates
- Smooth user experience

### Manual Refresh
- Press 'r' to refresh all data
- Updates nodes, channels, and config
- Shows "Data refreshed" in status bar
- Timestamp updates

## Error Handling

### Connection Errors
```
Connect to Meshtastic Device
═════════════════════════════════════════════════════════════════════════

Enter the port or IP address of your Meshtastic device:

> /dev/ttyUSB0

Error: failed to connect to radio: no such file or directory

Examples:
  • Serial: /dev/ttyUSB0, /dev/cu.SLAB_USBtoUART, COM3
  • TCP: 192.168.1.100, meshtastic.local

Press Enter to connect
```

### General Errors
- Shown in status bar
- Red color for visibility
- Clear error messages
- Non-intrusive

## Usage Examples

### Example 1: Monitoring Messages
1. Connect to device
2. Press F1 to go to Messages tab
3. Watch messages appear in real-time
4. Type a message and press Enter to send

### Example 2: Checking Node Status
1. Connect to device
2. Press F2 to go to Nodes tab
3. View all nodes and their battery levels
4. Press 'r' to refresh if needed

### Example 3: Viewing Configuration
1. Connect to device
2. Press F4 to go to Config tab
3. Review radio settings
4. Check LoRa parameters

## Tips

1. **Quick Navigation**: Use F1-F5 for instant tab switching
2. **Stay Updated**: Messages auto-refresh, but press 'r' for other data
3. **Help Always Available**: Press F5 anytime for keyboard shortcuts
4. **Broadcast Messages**: Messages sent from TUI go to all nodes
5. **Terminal Size**: Larger terminal = more content visible

## Future Enhancements

Potential visual improvements:
- Scrollable message history
- Node detail popup
- Configuration editing forms
- Message filtering UI
- Custom color themes
- Progress indicators
- Notification badges

