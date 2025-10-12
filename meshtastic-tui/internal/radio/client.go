package radio

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/lmatte7/meshtastic-tui/internal/db"
)

var logger *log.Logger

func init() {
	// Create logger that writes to ./tmp.log
	logFile, err := os.OpenFile("./tmp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		logger = log.New(os.Stderr, "[RADIO] ", log.LstdFlags|log.Lshortfile)
	} else {
		logger = log.New(logFile, "[RADIO] ", log.LstdFlags|log.Lshortfile)
	}
}

// Client wraps the gomesh.Radio with additional functionality for the TUI
type Client struct {
	radio        gomesh.Radio
	connected    bool
	port         string
	db           *db.DB
	currentRadio *db.Radio // Current radio we're connected to
}

// NewClient creates a new radio client
func NewClient() (*Client, error) {
	database, err := db.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Client{
		connected: false,
		db:        database,
	}, nil
}

// Connect establishes a connection to the radio
func (c *Client) Connect(port string) error {
	logger.Printf("Starting connection to port: %s", port)
	c.port = port

	logger.Printf("Calling radio.Init(%s)", port)
	err := c.radio.Init(port)
	if err != nil {
		logger.Printf("radio.Init failed: %v", err)
		c.connected = false
		return fmt.Errorf("failed to connect to radio: %w", err)
	}
	logger.Printf("radio.Init succeeded")
	c.connected = true

	// Identify the radio and store in database
	logger.Printf("Starting radio identification")
	if err := c.identifyRadio(); err != nil {
		logger.Printf("identifyRadio failed: %v", err)
		return fmt.Errorf("failed to identify radio: %w", err)
	}
	logger.Printf("Radio identification completed")

	// Load initial data from radio and cache in database
	logger.Printf("Starting initial data sync")
	if err := c.syncFromRadio(); err != nil {
		logger.Printf("syncFromRadio failed: %v", err)
		return fmt.Errorf("failed to sync initial data: %w", err)
	}
	logger.Printf("Initial data sync completed")

	logger.Printf("Connection fully established")
	return nil
}

// Disconnect closes the connection to the radio
func (c *Client) Disconnect() error {
	if c.connected {
		c.radio.Close()
		c.connected = false
		c.currentRadio = nil
	}
	if c.db != nil {
		c.db.Close()
	}
	return nil
}

// identifyRadio identifies the connected radio and stores it in database
func (c *Client) identifyRadio() error {
	logger.Printf("identifyRadio: Starting radio identification")

	// Try to get radio info with retries and better error handling
	var responses []*gomeshproto.FromRadio
	var err error

	// Try multiple times as the radio might need time to respond
	for i := 0; i < 3; i++ {
		logger.Printf("identifyRadio: Attempt %d/3 to get radio info", i+1)
		responses, err = c.radio.GetRadioInfo()
		logger.Printf("identifyRadio: GetRadioInfo returned %d responses, error: %v", len(responses), err)

		if err == nil && len(responses) > 0 {
			logger.Printf("identifyRadio: Successfully got radio info on attempt %d", i+1)
			break
		}
		if i < 2 { // Don't sleep on last attempt
			logger.Printf("identifyRadio: Sleeping 500ms before retry")
			time.Sleep(time.Millisecond * 500)
		}
	}

	if err != nil {
		logger.Printf("identifyRadio: All attempts failed, creating placeholder radio. Error: %v", err)
		// If we can't get radio info, create a placeholder radio entry
		// Use a default node ID based on port for now
		defaultNodeID := uint32(12345) // Placeholder
		radio, dbErr := c.db.GetOrCreateRadio(defaultNodeID, "Unknown Radio", "Unknown", c.port)
		if dbErr != nil {
			logger.Printf("identifyRadio: Failed to create placeholder radio: %v", dbErr)
			return fmt.Errorf("failed to create placeholder radio: %w", dbErr)
		}
		c.currentRadio = radio
		logger.Printf("identifyRadio: Created placeholder radio with ID %d", defaultNodeID)
		return nil // Don't fail connection for this
	}

	// Find our own node info
	var myNodeID uint32
	var longName, shortName string

	logger.Printf("identifyRadio: Parsing %d responses to find node info", len(responses))
	for i, response := range responses {
		logger.Printf("identifyRadio: Processing response %d", i)
		if response == nil {
			logger.Printf("identifyRadio: Response %d is nil, skipping", i)
			continue
		}

		// Safely check the payload variant
		payloadVariant := response.GetPayloadVariant()
		if payloadVariant == nil {
			logger.Printf("identifyRadio: Response %d has nil payload variant, skipping", i)
			continue
		}

		logger.Printf("identifyRadio: Response %d payload variant type: %T", i, payloadVariant)

		if nodeInfo, ok := payloadVariant.(*gomeshproto.FromRadio_NodeInfo); ok {
			logger.Printf("identifyRadio: Found NodeInfo in response %d", i)
			if nodeInfo.NodeInfo != nil {
				myNodeID = nodeInfo.NodeInfo.Num
				logger.Printf("identifyRadio: Found node ID: %d", myNodeID)
				if nodeInfo.NodeInfo.User != nil {
					longName = nodeInfo.NodeInfo.User.LongName
					shortName = nodeInfo.NodeInfo.User.ShortName
					logger.Printf("identifyRadio: Found names - Long: '%s', Short: '%s'", longName, shortName)
				} else {
					logger.Printf("identifyRadio: NodeInfo.User is nil")
				}
				// Take the first valid node we find
				if myNodeID != 0 {
					logger.Printf("identifyRadio: Using node ID %d", myNodeID)
					break
				}
			} else {
				logger.Printf("identifyRadio: NodeInfo is nil in response %d", i)
			}
		}
	}

	// If we still don't have a node ID, use a default
	if myNodeID == 0 {
		myNodeID = uint32(12345) // Placeholder
		longName = "Unknown Radio"
		shortName = "Unknown"
	}

	// Create or get radio record
	radio, err := c.db.GetOrCreateRadio(myNodeID, longName, shortName, c.port)
	if err != nil {
		return fmt.Errorf("failed to create/get radio record: %w", err)
	}

	c.currentRadio = radio
	return nil
}

// syncFromRadio loads data from radio and stores in database
func (c *Client) syncFromRadio() error {
	logger.Printf("syncFromRadio: Starting data sync")
	if c.currentRadio == nil {
		logger.Printf("syncFromRadio: No current radio, aborting")
		return fmt.Errorf("radio not identified")
	}

	// Sync nodes (don't fail connection if this fails)
	logger.Printf("syncFromRadio: Syncing nodes")
	if err := c.syncNodes(); err != nil {
		logger.Printf("syncFromRadio: Failed to sync nodes: %v", err)
		fmt.Printf("Warning: failed to sync nodes: %v\n", err)
	} else {
		logger.Printf("syncFromRadio: Nodes synced successfully")
	}

	// Skip channels and config sync for now due to gomesh library wire-format issues
	logger.Printf("syncFromRadio: Skipping channel sync (known wire-format issue)")
	logger.Printf("syncFromRadio: Skipping config sync (known wire-format issue)")

	logger.Printf("syncFromRadio: Data sync completed")
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.connected
}

// GetPort returns the current port
func (c *Client) GetPort() string {
	return c.port
}

// GetRadioInfo retrieves radio information
func (c *Client) GetRadioInfo() ([]*gomeshproto.FromRadio, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to radio")
	}
	return c.radio.GetRadioInfo()
}

// GetChannels retrieves channel information
func (c *Client) GetChannels() ([]*gomeshproto.Channel, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to radio")
	}
	return c.radio.GetChannels()
}

// GetRadioConfig retrieves radio configuration
func (c *Client) GetRadioConfig() ([]*gomeshproto.FromRadio_Config, []*gomeshproto.FromRadio_ModuleConfig, error) {
	if !c.connected {
		return nil, nil, fmt.Errorf("not connected to radio")
	}
	return c.radio.GetRadioConfig()
}

// SendTextMessage sends a text message
func (c *Client) SendTextMessage(message string, to int64, channel int64) error {
	if !c.connected {
		return fmt.Errorf("not connected to radio")
	}
	return c.radio.SendTextMessage(message, to, channel)
}

// ReadResponse reads responses from the radio
func (c *Client) ReadResponse(wait bool) ([]*gomeshproto.FromRadio, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to radio")
	}
	return c.radio.ReadResponse(wait)
}

// SetLocation sets the GPS location
func (c *Client) SetLocation(lat, long, alt int32) error {
	if !c.connected {
		return fmt.Errorf("not connected to radio")
	}
	return c.radio.SetLocation(lat, long, alt)
}

// SetRadioOwner sets the radio owner name
func (c *Client) SetRadioOwner(name string) error {
	if !c.connected {
		return fmt.Errorf("not connected to radio")
	}
	return c.radio.SetRadioOwner(name)
}

// SetRadioConfig sets a radio configuration value
func (c *Client) SetRadioConfig(key, value string) error {
	if !c.connected {
		return fmt.Errorf("not connected to radio")
	}
	return c.radio.SetRadioConfig(key, value)
}

// Message represents a received message
type Message struct {
	From      uint32
	To        uint32
	Channel   uint32
	Payload   string
	Timestamp time.Time
	PortNum   string
}

// ExtractMessages extracts text messages from radio responses
func ExtractMessages(responses []*gomeshproto.FromRadio) []Message {
	messages := []Message{}
	for _, response := range responses {
		if packet, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_Packet); ok {
			if packet.Packet.GetDecoded().GetPortnum() == gomeshproto.PortNum_TEXT_MESSAGE_APP {
				messages = append(messages, Message{
					From:      packet.Packet.From,
					To:        packet.Packet.To,
					Channel:   packet.Packet.Channel,
					Payload:   string(packet.Packet.GetDecoded().Payload),
					Timestamp: time.Now(),
					PortNum:   packet.Packet.GetDecoded().GetPortnum().String(),
				})
			}
		}
	}
	return messages
}

// Node represents a mesh node
type Node struct {
	Num          uint32
	LongName     string
	ShortName    string
	BatteryLevel uint32
	Voltage      float32
	Altitude     int32
	Latitude     int32
	Longitude    int32
	ChannelUtil  float32
	AirUtilTx    float32
	LastHeard    time.Time
}

// ExtractNodes extracts node information from radio responses
func ExtractNodes(responses []*gomeshproto.FromRadio) []Node {
	logger.Printf("ExtractNodes: Processing %d responses", len(responses))
	nodes := []Node{}

	for i, response := range responses {
		logger.Printf("ExtractNodes: Processing response %d", i)
		if response == nil {
			logger.Printf("ExtractNodes: Response %d is nil, skipping", i)
			continue
		}

		payloadVariant := response.GetPayloadVariant()
		if payloadVariant == nil {
			logger.Printf("ExtractNodes: Response %d has nil payload variant, skipping", i)
			continue
		}

		logger.Printf("ExtractNodes: Response %d payload variant type: %T", i, payloadVariant)

		if nodeInfo, ok := payloadVariant.(*gomeshproto.FromRadio_NodeInfo); ok {
			logger.Printf("ExtractNodes: Found NodeInfo in response %d", i)
			if nodeInfo.NodeInfo != nil {
				node := Node{
					Num: nodeInfo.NodeInfo.Num,
				}
				logger.Printf("ExtractNodes: Creating node with ID %d", node.Num)

				if nodeInfo.NodeInfo.User != nil {
					node.LongName = nodeInfo.NodeInfo.User.LongName
					node.ShortName = nodeInfo.NodeInfo.User.ShortName
					logger.Printf("ExtractNodes: Node %d names - Long: '%s', Short: '%s'",
						node.Num, node.LongName, node.ShortName)
				} else {
					logger.Printf("ExtractNodes: Node %d has no User info", node.Num)
				}

				if nodeInfo.NodeInfo.DeviceMetrics != nil {
					node.BatteryLevel = nodeInfo.NodeInfo.DeviceMetrics.BatteryLevel
					node.Voltage = nodeInfo.NodeInfo.DeviceMetrics.Voltage
					node.ChannelUtil = nodeInfo.NodeInfo.DeviceMetrics.ChannelUtilization
					node.AirUtilTx = nodeInfo.NodeInfo.DeviceMetrics.AirUtilTx
					logger.Printf("ExtractNodes: Node %d metrics - Battery: %d%%, Voltage: %.2fV",
						node.Num, node.BatteryLevel, node.Voltage)
				} else {
					logger.Printf("ExtractNodes: Node %d has no DeviceMetrics", node.Num)
				}

				if nodeInfo.NodeInfo.Position != nil {
					node.Altitude = nodeInfo.NodeInfo.Position.Altitude
					node.Latitude = nodeInfo.NodeInfo.Position.LatitudeI
					node.Longitude = nodeInfo.NodeInfo.Position.LongitudeI
					logger.Printf("ExtractNodes: Node %d position - Alt: %d, Lat: %d, Lon: %d",
						node.Num, node.Altitude, node.Latitude, node.Longitude)
				} else {
					logger.Printf("ExtractNodes: Node %d has no Position info", node.Num)
				}

				nodes = append(nodes, node)
				logger.Printf("ExtractNodes: Added node %d to list", node.Num)
			} else {
				logger.Printf("ExtractNodes: NodeInfo is nil in response %d", i)
			}
		}
	}

	logger.Printf("ExtractNodes: Extracted %d nodes total", len(nodes))
	return nodes
}

// syncNodes syncs node data from radio to database
func (c *Client) syncNodes() error {
	logger.Printf("syncNodes: Getting radio info for nodes")
	responses, err := c.radio.GetRadioInfo()
	if err != nil {
		logger.Printf("syncNodes: GetRadioInfo failed: %v", err)
		return err
	}
	logger.Printf("syncNodes: Got %d responses", len(responses))

	// Extract nodes safely
	logger.Printf("syncNodes: Extracting nodes from responses")
	nodes := ExtractNodes(responses)
	logger.Printf("syncNodes: Extracted %d nodes", len(nodes))

	if len(nodes) == 0 {
		logger.Printf("syncNodes: No nodes to sync")
		return nil // No nodes to sync, not an error
	}

	for i, node := range nodes {
		logger.Printf("syncNodes: Processing node %d: ID=%d, LongName='%s', ShortName='%s'",
			i, node.Num, node.LongName, node.ShortName)

		dbNode := &db.Node{
			NodeID:       node.Num,
			LongName:     node.LongName,
			ShortName:    node.ShortName,
			BatteryLevel: node.BatteryLevel,
			Voltage:      node.Voltage,
			Altitude:     node.Altitude,
			Latitude:     node.Latitude,
			Longitude:    node.Longitude,
			ChannelUtil:  node.ChannelUtil,
			AirUtilTx:    node.AirUtilTx,
			LastHeard:    node.LastHeard,
		}

		if err := c.db.UpsertNode(c.currentRadio.ID, dbNode); err != nil {
			// Log error but continue with other nodes
			logger.Printf("syncNodes: Failed to upsert node %d: %v", node.Num, err)
			fmt.Printf("Warning: failed to upsert node %d: %v\n", node.Num, err)
		} else {
			logger.Printf("syncNodes: Successfully upserted node %d", node.Num)
		}
	}
	logger.Printf("syncNodes: Completed processing %d nodes", len(nodes))
	return nil
}

// syncChannels syncs channel data from radio to database
func (c *Client) syncChannels() error {
	channels, err := c.radio.GetChannels()
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		return nil // No channels to sync, not an error
	}

	for _, ch := range channels {
		if ch == nil || ch.Settings == nil {
			continue // Skip invalid channels
		}

		dbChannel := &db.Channel{
			Index:    uint32(ch.Index),
			Name:     ch.Settings.Name,
			PSK:      string(ch.Settings.Psk),
			Uplink:   ch.Settings.UplinkEnabled,
			Downlink: ch.Settings.DownlinkEnabled,
		}

		// Determine role safely
		switch ch.Role {
		case gomeshproto.Channel_PRIMARY:
			dbChannel.Role = "PRIMARY"
		case gomeshproto.Channel_SECONDARY:
			dbChannel.Role = "SECONDARY"
		default:
			dbChannel.Role = "DISABLED"
		}

		if err := c.db.UpsertChannel(c.currentRadio.ID, dbChannel); err != nil {
			// Log error but continue with other channels
			fmt.Printf("Warning: failed to upsert channel %d: %v\n", ch.Index, err)
		}
	}
	return nil
}

// syncConfig syncs config data from radio to database
func (c *Client) syncConfig() error {
	configs, moduleConfigs, err := c.radio.GetRadioConfig()
	if err != nil {
		return err
	}

	// Store basic config info (simplified for now)
	for i, cfg := range configs {
		key := fmt.Sprintf("config_%d", i)
		value := fmt.Sprintf("%T", cfg)
		if err := c.db.UpsertConfig(c.currentRadio.ID, key, value, "device"); err != nil {
			return fmt.Errorf("failed to upsert config %s: %w", key, err)
		}
	}

	for i, modCfg := range moduleConfigs {
		key := fmt.Sprintf("module_config_%d", i)
		value := fmt.Sprintf("%T", modCfg)
		if err := c.db.UpsertConfig(c.currentRadio.ID, key, value, "module"); err != nil {
			return fmt.Errorf("failed to upsert module config %s: %w", key, err)
		}
	}

	return nil
}

// GetCurrentRadio returns the current radio info
func (c *Client) GetCurrentRadio() *db.Radio {
	return c.currentRadio
}

// GetNodesFromDB gets nodes from database
func (c *Client) GetNodesFromDB() ([]*db.Node, error) {
	if c.currentRadio == nil {
		return nil, fmt.Errorf("no radio connected")
	}
	return c.db.GetNodes(c.currentRadio.ID)
}

// GetChannelsFromDB gets channels from database
func (c *Client) GetChannelsFromDB() ([]*db.Channel, error) {
	if c.currentRadio == nil {
		return nil, fmt.Errorf("no radio connected")
	}
	return c.db.GetChannels(c.currentRadio.ID)
}

// GetMessagesFromDB gets recent messages from database
func (c *Client) GetMessagesFromDB(limit int) ([]*db.Message, error) {
	if c.currentRadio == nil {
		return nil, fmt.Errorf("no radio connected")
	}
	return c.db.GetRecentMessages(c.currentRadio.ID, limit)
}

// SendTextMessageWithDB sends a message and stores it in database
func (c *Client) SendTextMessageWithDB(message string, to int64, channel int64) error {
	if !c.connected || c.currentRadio == nil {
		return fmt.Errorf("not connected to radio")
	}

	// Store message in database as pending
	dbMessage := &db.Message{
		FromNodeID: c.currentRadio.NodeID,
		ToNodeID:   uint32(to),
		ChannelID:  uint32(channel),
		Content:    message,
		Status:     db.MessageStatusPending,
		Direction:  db.MessageDirectionOutgoing,
		PortNum:    "TEXT_MESSAGE_APP",
		Timestamp:  time.Now(),
	}

	messageID, err := c.db.InsertMessage(c.currentRadio.ID, dbMessage)
	if err != nil {
		return fmt.Errorf("failed to store message in database: %w", err)
	}

	// Send to radio
	err = c.radio.SendTextMessage(message, to, channel)
	if err != nil {
		// Update message status to failed
		c.db.UpdateMessageStatus(messageID, db.MessageStatusFailed)
		return fmt.Errorf("failed to send message to radio: %w", err)
	}

	// For now, immediately mark as acknowledged
	// TODO: In the future, we should wait for actual radio acknowledgment
	c.db.UpdateMessageStatus(messageID, db.MessageStatusAcknowledged)

	return nil
}
