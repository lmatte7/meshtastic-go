package db

import (
	"time"
)

// Radio represents a connected radio device
type Radio struct {
	ID          int       `db:"id"`
	NodeID      uint32    `db:"node_id"`      // The radio's own node ID
	LongName    string    `db:"long_name"`    // Radio's long name
	ShortName   string    `db:"short_name"`   // Radio's short name
	HardwareID  string    `db:"hardware_id"`  // Hardware identifier
	LastSeen    time.Time `db:"last_seen"`    // Last connection time
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Node represents a mesh node
type Node struct {
	ID           int       `db:"id"`
	RadioID      int       `db:"radio_id"`      // Foreign key to Radio
	NodeID       uint32    `db:"node_id"`       // Meshtastic node ID
	LongName     string    `db:"long_name"`
	ShortName    string    `db:"short_name"`
	BatteryLevel uint32    `db:"battery_level"`
	Voltage      float32   `db:"voltage"`
	Altitude     int32     `db:"altitude"`
	Latitude     int32     `db:"latitude"`
	Longitude    int32     `db:"longitude"`
	ChannelUtil  float32   `db:"channel_util"`
	AirUtilTx    float32   `db:"air_util_tx"`
	LastHeard    time.Time `db:"last_heard"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// Channel represents a mesh channel
type Channel struct {
	ID        int       `db:"id"`
	RadioID   int       `db:"radio_id"`   // Foreign key to Radio
	Index     uint32    `db:"index"`      // Channel index
	Name      string    `db:"name"`
	Role      string    `db:"role"`
	PSK       string    `db:"psk"`        // Pre-shared key
	Uplink    bool      `db:"uplink"`
	Downlink  bool      `db:"downlink"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Message represents a mesh message
type Message struct {
	ID          int                `db:"id"`
	RadioID     int                `db:"radio_id"`     // Foreign key to Radio
	FromNodeID  uint32             `db:"from_node_id"` // Sender node ID
	ToNodeID    uint32             `db:"to_node_id"`   // Recipient node ID (0 = broadcast)
	ChannelID   uint32             `db:"channel_id"`   // Channel index
	Content     string             `db:"content"`      // Message text
	Status      MessageStatus      `db:"status"`       // Message status
	Direction   MessageDirection   `db:"direction"`    // Incoming or outgoing
	PortNum     string             `db:"port_num"`     // Meshtastic port number
	Timestamp   time.Time          `db:"timestamp"`    // When message was sent/received
	CreatedAt   time.Time          `db:"created_at"`
	UpdatedAt   time.Time          `db:"updated_at"`
}

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusPending     MessageStatus = "pending"     // Sent to radio, waiting for ack
	MessageStatusAcknowledged MessageStatus = "acknowledged" // Confirmed by radio
	MessageStatusFailed      MessageStatus = "failed"      // Failed to send
	MessageStatusReceived    MessageStatus = "received"    // Incoming message
)

// MessageDirection represents the direction of a message
type MessageDirection string

const (
	MessageDirectionIncoming MessageDirection = "incoming"
	MessageDirectionOutgoing MessageDirection = "outgoing"
)

// Config represents radio configuration
type Config struct {
	ID        int       `db:"id"`
	RadioID   int       `db:"radio_id"` // Foreign key to Radio
	Key       string    `db:"key"`      // Config key
	Value     string    `db:"value"`    // Config value
	Category  string    `db:"category"` // Config category (device, lora, etc.)
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
