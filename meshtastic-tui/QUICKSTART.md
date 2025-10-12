# Meshtastic TUI - Quick Start Guide

## Installation

```bash
cd meshtastic-tui
go build
./meshtastic-tui
```

## First Time Setup

1. **Launch the TUI**
   ```bash
   ./meshtastic-tui
   ```

2. **Connect to Your Device**
   - Enter your device port or IP address
   - Examples:
     - Serial: `/dev/ttyUSB0`, `/dev/ttyACM0`, `COM3`
     - TCP/IP: `192.168.1.100`
   - Press **Enter** to connect

3. **Wait for Connection**
   - Status will show "● Connected" in green when ready
   - TUI will automatically switch to Messages view
   - Data will be refreshed automatically

## Essential Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| **F1** | Messages view |
| **F2** | Nodes view |
| **F3** | Channels view |
| **F4** | Configuration view |
| **F5** | Help screen |
| **R** | Refresh data from device (when not typing) |
| **Q** | Quit application |
| **Ctrl+C** | Quit application |

### Messages View
| Key | Action |
|-----|--------|
| **Tab** | Next input field (Message → To → Channel) |
| **Shift+Tab** | Previous input field |
| **Enter** | Send message |
| **↑ / ↓** | Scroll message history |
| **PgUp / PgDn** | Scroll page up/down |

### Scrolling (All Views)
| Key | Action |
|-----|--------|
| **↑ / ↓** | Scroll up/down one line |
| **PgUp / PgDn** | Scroll up/down one page |
| **Home** | Jump to top |
| **End** | Jump to bottom |

## Common Tasks

### Send a Broadcast Message
1. Press **F1** (Messages view)
2. Type your message in the **Message** field
3. Leave **To** field empty (or enter `0`)
4. Leave **Channel** field empty (or enter `0`)
5. Press **Enter**

**Result**: Message sent to all nodes on default channel

### Send a Direct Message
1. Press **F1** (Messages view)
2. Type your message in the **Message** field
3. Press **Tab** to move to **To** field
4. Enter the node ID (e.g., `123456789`)
5. Press **Enter**

**Result**: Message sent only to specified node

### Send on Specific Channel
1. Press **F1** (Messages view)
2. Type your message in the **Message** field
3. Press **Tab** twice to move to **Channel** field
4. Enter channel number (e.g., `1`, `2`, `3`)
5. Press **Enter**

**Result**: Message sent on specified channel

### View All Nodes
1. Press **F2** (Nodes view)
2. Use **↑/↓** to scroll through the list
3. View node ID, name, battery, voltage, GPS coordinates

### View Channel Configuration
1. Press **F3** (Channels view)
2. Use **↑/↓** to scroll through channels
3. View channel index, name, role, uplink/downlink status, PSK

### View Device Configuration
1. Press **F4** (Config view)
2. Use **↑/↓** to scroll through settings
3. Press **R** to refresh if data is stale

### Get Help
1. Press **F5** (Help view)
2. View all keyboard shortcuts and features
3. Press **F1-F4** to return to other views

## Message Field Reference

### Message Field
- **Purpose**: The text content of your message
- **Max Length**: 237 characters
- **Required**: Yes
- **Example**: `Hello mesh network!`

### To Field
- **Purpose**: Recipient node ID
- **Values**:
  - `0` or empty = Broadcast to all nodes
  - Node ID = Send to specific node (e.g., `123456789`)
- **Required**: No (defaults to broadcast)
- **Example**: `123456789`

### Channel Field
- **Purpose**: Which channel to send on
- **Values**:
  - `0` or empty = Default channel
  - Channel number = Specific channel (e.g., `1`, `2`, `3`)
- **Required**: No (defaults to channel 0)
- **Example**: `1`

## Tips & Tricks

### Efficient Navigation
- Use **F-keys** for instant view switching
- Use **Tab** to quickly move between message fields
- Use **PgUp/PgDn** for fast scrolling

### Message Composition
- Type your message first, then Tab to other fields
- Leave To/Channel empty for quick broadcasts
- Press **Enter** from any field to send

### Monitoring
- Messages auto-refresh every second
- Press **R** to manually refresh nodes/channels/config
- Status bar shows last action and timestamp

### Troubleshooting
- If connection fails, check device port/IP
- If no data appears, press **R** to refresh
- If messages aren't sending, check connection status (top right)
- Press **F5** for help anytime

## Status Indicators

### Connection Status (Top Right)
- **● Connected** (green) = Device connected and ready
- **○ Disconnected** (red) = No device connection

### Input Focus (Messages View)
- **Highlighted border** (green) = Field has focus
- **Normal border** (cyan) = Field not focused
- Use **Tab** to change focus

### Status Bar (Bottom)
- Shows context-appropriate keyboard shortcuts
- Displays last action and timestamp
- Updates automatically

## View Descriptions

### Connect View
- Enter device connection details
- Shows connection status and errors
- Provides example connection strings

### Messages View (F1)
- Scrollable message history
- Three input fields for composing messages
- Real-time message updates (every second)
- Shows sender, recipient, timestamp for each message

### Nodes View (F2)
- Table of all mesh nodes
- Shows node ID, name, battery, voltage, GPS coordinates
- Scrollable list
- Updates when you press **R**

### Channels View (F3)
- Table of all configured channels
- Shows index, name, role, uplink/downlink status
- Displays PSK (encryption key) if available
- Scrollable list

### Config View (F4)
- Device configuration settings
- Module configuration
- Scrollable content
- Press **R** to refresh

### Help View (F5)
- Complete keyboard shortcut reference
- Feature descriptions
- Always accessible from any view

## Example Workflow

### Typical Session
1. **Launch**: `./meshtastic-tui`
2. **Connect**: Enter `/dev/ttyUSB0`, press Enter
3. **Wait**: See "● Connected" status
4. **Monitor**: Watch messages arrive in real-time
5. **Send**: Type message, press Enter
6. **Check Nodes**: Press F2 to see who's online
7. **View Channels**: Press F3 to see channel config
8. **Quit**: Press Q when done

### Advanced Usage
1. **Direct Message**: F1 → Type message → Tab → Enter node ID → Enter
2. **Channel Message**: F1 → Type message → Tab → Tab → Enter channel → Enter
3. **Monitor Specific**: F2 to see nodes, note ID, F1 to message them
4. **Configuration**: F4 to view settings, R to refresh

## Troubleshooting

### Can't Connect
- Check device is plugged in
- Verify correct port/IP address
- Try different port (e.g., `/dev/ttyACM0` instead of `/dev/ttyUSB0`)
- Check device permissions: `sudo chmod 666 /dev/ttyUSB0`

### No Messages Appearing
- Check connection status (top right)
- Press **R** to refresh
- Verify device is on mesh network
- Check if other nodes are transmitting

### Can't Send Messages
- Verify "● Connected" status
- Check message field isn't empty
- Try broadcast first (leave To/Channel empty)
- Check device has power and antenna

### Display Issues
- Resize terminal window (minimum 80x24 recommended)
- Check terminal supports colors
- Try different terminal emulator

### Performance Issues
- Long message history can slow down
- Restart TUI to clear message buffer
- Consider reducing auto-refresh frequency (code modification)

## Getting Help

### In-App Help
- Press **F5** anytime for keyboard shortcuts
- Status bar shows context-appropriate help
- Error messages appear in status bar

### Documentation
- `README.md` - Full documentation
- `ARCHITECTURE.md` - Technical details
- `FIXES.md` - Recent improvements
- `QUICKSTART.md` - This file

### Support
- Check GitHub issues
- Review meshtastic-go documentation
- Consult Meshtastic community forums

## Next Steps

Once you're comfortable with the basics:
1. Experiment with direct messages to specific nodes
2. Try sending on different channels
3. Monitor node battery levels and GPS positions
4. Explore device configuration options
5. Customize the TUI (see ARCHITECTURE.md)

Happy meshing! 📡

