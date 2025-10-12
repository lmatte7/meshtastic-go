# Meshtastic TUI Architecture

## Overview

This document describes the architecture and design decisions for the Meshtastic TUI application.

## Design Principles

### 1. Separation from meshtastic-go
The TUI is built as a **standalone application** that imports meshtastic-go as a library dependency. This design ensures:
- No modifications to the original meshtastic-go codebase
- Clean separation of concerns
- Easy maintenance and updates
- The TUI can be developed independently

### 2. Modern TUI Framework
Built using [Bubble Tea](https://github.com/charmbracelet/bubbletea), a modern Go framework for building terminal applications based on The Elm Architecture:
- **Model**: Application state
- **Update**: State transitions based on messages
- **View**: Rendering the UI

## Project Structure

```
meshtastic-tui/
├── main.go                    # Application entry point
├── go.mod                     # Go module with dependencies
├── internal/
│   ├── radio/
│   │   └── client.go         # Radio client wrapper
│   └── ui/
│       ├── model.go          # Main TUI model and state
│       ├── commands.go       # Async operations
│       └── views.go          # View rendering functions
├── README.md                  # User documentation
├── ARCHITECTURE.md            # This file
├── build.sh                   # Build script
└── .gitignore                # Git ignore rules
```

## Components

### Main Entry Point (`main.go`)

Simple entry point that:
1. Creates a new Bubble Tea program
2. Initializes the UI model
3. Enables alternate screen buffer and mouse support
4. Runs the program

### Radio Client (`internal/radio/client.go`)

Wraps the `gomesh.Radio` type with additional functionality:

**Purpose:**
- Provide a clean interface to the meshtastic-go library
- Manage connection state
- Extract and format data from protobuf responses

**Key Types:**
- `Client`: Wrapper around `gomesh.Radio`
- `Message`: Simplified message representation
- `Node`: Simplified node information

**Key Methods:**
- `Connect(port)`: Establish connection
- `Disconnect()`: Close connection
- `GetRadioInfo()`: Retrieve radio information
- `SendTextMessage()`: Send messages
- `ReadResponse()`: Read incoming data

**Helper Functions:**
- `ExtractMessages()`: Parse messages from responses
- `ExtractNodes()`: Parse node information

### UI Model (`internal/ui/model.go`)

The main application state and logic:

**State:**
- `client`: Radio client instance
- `currentTab`: Active tab (Connect, Messages, Nodes, etc.)
- `width`, `height`: Terminal dimensions
- Connection state (input, connecting flag, errors)
- Messages state (history, input field)
- Nodes, channels, config data

**Tabs:**
- `TabConnect`: Connection management
- `TabMessages`: Message viewing and sending
- `TabNodes`: Node list and information
- `TabChannels`: Channel configuration
- `TabConfig`: Radio settings
- `TabHelp`: Help and shortcuts

**Key Methods:**
- `Init()`: Initialize the model
- `Update(msg)`: Handle messages and update state
- `View()`: Render the UI
- `handleEnter()`: Process Enter key based on context

### Commands (`internal/ui/commands.go`)

Async operations that don't block the UI:

**Message Types:**
- `tickMsg`: Periodic timer for auto-refresh
- `connectSuccessMsg`: Connection established
- `connectErrorMsg`: Connection failed
- `dataRefreshMsg`: Data loaded from radio
- `newMessagesMsg`: New messages received
- `messageSentMsg`: Message sent successfully
- `errorMsg`: General error

**Commands:**
- `tickCmd()`: Timer for periodic updates
- `connectCmd()`: Connect to radio
- `refreshDataCmd()`: Load all data
- `checkMessagesCmd()`: Check for new messages
- `sendMessageCmd()`: Send a message

**Helper Functions:**
- `extractChannels()`: Parse channel data
- `formatConfig()`: Format configuration for display

### Views (`internal/ui/views.go`)

Rendering functions for each screen:

**Styling:**
- Color scheme with primary, secondary, accent colors
- Consistent styles for headers, tables, inputs
- Lipgloss for layout and styling

**View Functions:**
- `renderHeader()`: Title bar with tabs and status
- `renderFooter()`: Status bar with help text
- `renderConnect()`: Connection screen
- `renderMessages()`: Message list and input
- `renderNodes()`: Node table
- `renderChannels()`: Channel table
- `renderConfig()`: Configuration display
- `renderHelp()`: Help screen

## Data Flow

### Connection Flow
```
User enters port → Enter key → connectCmd() → 
  Radio.Init() → connectSuccessMsg → 
  Switch to Messages tab → refreshDataCmd()
```

### Message Sending Flow
```
User types message → Enter key → sendMessageCmd() →
  Radio.SendTextMessage() → messageSentMsg →
  Clear input field
```

### Message Receiving Flow
```
tickMsg (every second) → checkMessagesCmd() →
  Radio.ReadResponse() → ExtractMessages() →
  newMessagesMsg → Append to message list
```

### Data Refresh Flow
```
User presses 'r' → refreshDataCmd() →
  GetRadioInfo() + GetChannels() + GetRadioConfig() →
  Extract and format data → dataRefreshMsg →
  Update state
```

## Key Design Decisions

### 1. Tab-Based Navigation
- Organized by functionality
- Easy keyboard navigation (Tab, F1-F5)
- Clear visual indication of active tab

### 2. Real-Time Updates
- Messages auto-refresh every second
- Non-blocking async operations
- Responsive UI during data loading

### 3. Error Handling
- Connection errors shown in Connect tab
- General errors in status bar
- Graceful degradation (show "N/A" for missing data)

### 4. Minimal Dependencies
- Only essential libraries (Bubble Tea, Lipgloss, Bubbles)
- Direct use of meshtastic-go and gomesh
- No unnecessary abstractions

### 5. User Experience
- Intuitive keyboard shortcuts
- Visual feedback for all actions
- Help screen always accessible (F5)
- Status bar shows last action and time

## State Management

The application uses The Elm Architecture pattern:

1. **Model**: All state in one place
2. **Messages**: Events that trigger state changes
3. **Update**: Pure function that produces new state
4. **Commands**: Side effects that produce messages
5. **View**: Pure function that renders state

This makes the application:
- Predictable and testable
- Easy to reason about
- Naturally concurrent (commands run async)

## Extension Points

To add new features:

### New Tab
1. Add tab constant to `model.go`
2. Add case in `Update()` for navigation
3. Create `render<TabName>()` in `views.go`
4. Add to tab bar in `renderHeader()`

### New Command
1. Define message type in `commands.go`
2. Create command function
3. Handle message in `Update()`
4. Trigger command from appropriate place

### New Radio Operation
1. Add method to `Client` in `radio/client.go`
2. Create command in `commands.go`
3. Handle in `Update()`
4. Display in appropriate view

## Testing Strategy

While not implemented yet, the architecture supports:

### Unit Tests
- Test helper functions (ExtractMessages, formatConfig)
- Test state transitions in Update()
- Mock radio client for UI tests

### Integration Tests
- Test with mock radio responses
- Verify data flow through commands

### Manual Testing
- Test with actual Meshtastic device
- Verify all tabs and features
- Test error conditions

## Performance Considerations

### Efficient Rendering
- Only re-render on state changes
- Lipgloss caches styles
- Minimal string allocations

### Message History
- Keep last N messages (currently 10 shown)
- Could add pagination for more

### Auto-Refresh
- 1 second interval for messages
- Manual refresh for other data
- Could make configurable

## Future Enhancements

Possible improvements:

1. **Configuration Editing**: Allow changing settings from TUI
2. **Message Filtering**: Filter by sender, channel, etc.
3. **Node Details**: Detailed view for selected node
4. **Message History**: Scrollable message list
5. **Direct Messages**: Send to specific node
6. **Channel Selection**: Choose channel for messages
7. **Logging**: Save messages to file
8. **Themes**: Customizable color schemes
9. **Plugins**: Extension system for custom features
10. **Performance Metrics**: Show radio statistics

## Dependencies

### Direct Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles` - UI components
- `github.com/lmatte7/meshtastic-go` - Meshtastic library (local)
- `github.com/lmatte7/gomesh` - Mesh protocol

### Why These Dependencies?
- **Bubble Tea**: Best-in-class Go TUI framework, well-maintained
- **Lipgloss**: Powerful styling, pairs with Bubble Tea
- **Bubbles**: Pre-built components (text input)
- **meshtastic-go**: Required for Meshtastic communication
- **gomesh**: Protocol implementation

## Conclusion

This architecture provides:
- Clean separation from meshtastic-go
- Modern, maintainable codebase
- Excellent user experience
- Easy to extend and modify
- Follows Go best practices

