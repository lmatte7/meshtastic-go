package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New() (*DB, error) {
	// Create data directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".meshtastic-tui")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database
	dbPath := filepath.Join(dataDir, "meshtastic.db")
	conn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates the database tables
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS radios (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_id INTEGER NOT NULL UNIQUE,
		long_name TEXT NOT NULL,
		short_name TEXT NOT NULL,
		hardware_id TEXT NOT NULL,
		last_seen DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS nodes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		radio_id INTEGER NOT NULL,
		node_id INTEGER NOT NULL,
		long_name TEXT NOT NULL DEFAULT '',
		short_name TEXT NOT NULL DEFAULT '',
		battery_level INTEGER NOT NULL DEFAULT 0,
		voltage REAL NOT NULL DEFAULT 0,
		altitude INTEGER NOT NULL DEFAULT 0,
		latitude INTEGER NOT NULL DEFAULT 0,
		longitude INTEGER NOT NULL DEFAULT 0,
		channel_util REAL NOT NULL DEFAULT 0,
		air_util_tx REAL NOT NULL DEFAULT 0,
		last_heard DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (radio_id) REFERENCES radios(id) ON DELETE CASCADE,
		UNIQUE(radio_id, node_id)
	);

	CREATE TABLE IF NOT EXISTS channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		radio_id INTEGER NOT NULL,
		index_num INTEGER NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT '',
		psk TEXT NOT NULL DEFAULT '',
		uplink BOOLEAN NOT NULL DEFAULT 0,
		downlink BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (radio_id) REFERENCES radios(id) ON DELETE CASCADE,
		UNIQUE(radio_id, index_num)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		radio_id INTEGER NOT NULL,
		from_node_id INTEGER NOT NULL,
		to_node_id INTEGER NOT NULL DEFAULT 0,
		channel_id INTEGER NOT NULL DEFAULT 0,
		content TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'received',
		direction TEXT NOT NULL DEFAULT 'incoming',
		port_num TEXT NOT NULL DEFAULT '',
		timestamp DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (radio_id) REFERENCES radios(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		radio_id INTEGER NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		category TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (radio_id) REFERENCES radios(id) ON DELETE CASCADE,
		UNIQUE(radio_id, key)
	);

	-- Indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_nodes_radio_id ON nodes(radio_id);
	CREATE INDEX IF NOT EXISTS idx_nodes_node_id ON nodes(node_id);
	CREATE INDEX IF NOT EXISTS idx_channels_radio_id ON channels(radio_id);
	CREATE INDEX IF NOT EXISTS idx_messages_radio_id ON messages(radio_id);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
	CREATE INDEX IF NOT EXISTS idx_messages_from_node ON messages(from_node_id);
	CREATE INDEX IF NOT EXISTS idx_messages_to_node ON messages(to_node_id);
	CREATE INDEX IF NOT EXISTS idx_configs_radio_id ON configs(radio_id);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// GetOrCreateRadio gets or creates a radio record
func (db *DB) GetOrCreateRadio(nodeID uint32, longName, shortName, hardwareID string) (*Radio, error) {
	// Try to get existing radio
	radio := &Radio{}
	err := db.conn.QueryRow(`
		SELECT id, node_id, long_name, short_name, hardware_id, last_seen, created_at, updated_at
		FROM radios WHERE node_id = ?
	`, nodeID).Scan(
		&radio.ID, &radio.NodeID, &radio.LongName, &radio.ShortName,
		&radio.HardwareID, &radio.LastSeen, &radio.CreatedAt, &radio.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new radio
		now := time.Now()
		result, err := db.conn.Exec(`
			INSERT INTO radios (node_id, long_name, short_name, hardware_id, last_seen, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, nodeID, longName, shortName, hardwareID, now, now, now)
		if err != nil {
			return nil, fmt.Errorf("failed to create radio: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get radio ID: %w", err)
		}

		radio.ID = int(id)
		radio.NodeID = nodeID
		radio.LongName = longName
		radio.ShortName = shortName
		radio.HardwareID = hardwareID
		radio.LastSeen = now
		radio.CreatedAt = now
		radio.UpdatedAt = now
	} else if err != nil {
		return nil, fmt.Errorf("failed to query radio: %w", err)
	} else {
		// Update last seen
		now := time.Now()
		_, err = db.conn.Exec(`
			UPDATE radios SET last_seen = ?, updated_at = ? WHERE id = ?
		`, now, now, radio.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update radio last seen: %w", err)
		}
		radio.LastSeen = now
		radio.UpdatedAt = now
	}

	return radio, nil
}

// GetRadioByID gets a radio by ID
func (db *DB) GetRadioByID(id int) (*Radio, error) {
	radio := &Radio{}
	err := db.conn.QueryRow(`
		SELECT id, node_id, long_name, short_name, hardware_id, last_seen, created_at, updated_at
		FROM radios WHERE id = ?
	`, id).Scan(
		&radio.ID, &radio.NodeID, &radio.LongName, &radio.ShortName,
		&radio.HardwareID, &radio.LastSeen, &radio.CreatedAt, &radio.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return radio, nil
}

// UpsertNode creates or updates a node
func (db *DB) UpsertNode(radioID int, node *Node) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		INSERT INTO nodes (
			radio_id, node_id, long_name, short_name, battery_level, voltage,
			altitude, latitude, longitude, channel_util, air_util_tx, last_heard,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(radio_id, node_id) DO UPDATE SET
			long_name = excluded.long_name,
			short_name = excluded.short_name,
			battery_level = excluded.battery_level,
			voltage = excluded.voltage,
			altitude = excluded.altitude,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			channel_util = excluded.channel_util,
			air_util_tx = excluded.air_util_tx,
			last_heard = excluded.last_heard,
			updated_at = excluded.updated_at
	`, radioID, node.NodeID, node.LongName, node.ShortName, node.BatteryLevel,
		node.Voltage, node.Altitude, node.Latitude, node.Longitude,
		node.ChannelUtil, node.AirUtilTx, node.LastHeard, now, now)
	return err
}

// GetNodes gets all nodes for a radio
func (db *DB) GetNodes(radioID int) ([]*Node, error) {
	rows, err := db.conn.Query(`
		SELECT id, radio_id, node_id, long_name, short_name, battery_level, voltage,
			   altitude, latitude, longitude, channel_util, air_util_tx, last_heard,
			   created_at, updated_at
		FROM nodes WHERE radio_id = ? ORDER BY long_name, short_name
	`, radioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		err := rows.Scan(
			&node.ID, &node.RadioID, &node.NodeID, &node.LongName, &node.ShortName,
			&node.BatteryLevel, &node.Voltage, &node.Altitude, &node.Latitude,
			&node.Longitude, &node.ChannelUtil, &node.AirUtilTx, &node.LastHeard,
			&node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

// UpsertChannel creates or updates a channel
func (db *DB) UpsertChannel(radioID int, channel *Channel) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		INSERT INTO channels (
			radio_id, index_num, name, role, psk, uplink, downlink, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(radio_id, index_num) DO UPDATE SET
			name = excluded.name,
			role = excluded.role,
			psk = excluded.psk,
			uplink = excluded.uplink,
			downlink = excluded.downlink,
			updated_at = excluded.updated_at
	`, radioID, channel.Index, channel.Name, channel.Role, channel.PSK,
		channel.Uplink, channel.Downlink, now, now)
	return err
}

// GetChannels gets all channels for a radio
func (db *DB) GetChannels(radioID int) ([]*Channel, error) {
	rows, err := db.conn.Query(`
		SELECT id, radio_id, index_num, name, role, psk, uplink, downlink, created_at, updated_at
		FROM channels WHERE radio_id = ? ORDER BY index_num
	`, radioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		err := rows.Scan(
			&channel.ID, &channel.RadioID, &channel.Index, &channel.Name,
			&channel.Role, &channel.PSK, &channel.Uplink, &channel.Downlink,
			&channel.CreatedAt, &channel.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, rows.Err()
}

// InsertMessage inserts a new message
func (db *DB) InsertMessage(radioID int, message *Message) (int64, error) {
	now := time.Now()
	result, err := db.conn.Exec(`
		INSERT INTO messages (
			radio_id, from_node_id, to_node_id, channel_id, content, status,
			direction, port_num, timestamp, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, radioID, message.FromNodeID, message.ToNodeID, message.ChannelID,
		message.Content, message.Status, message.Direction, message.PortNum,
		message.Timestamp, now, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateMessageStatus updates a message status
func (db *DB) UpdateMessageStatus(messageID int64, status MessageStatus) error {
	_, err := db.conn.Exec(`
		UPDATE messages SET status = ?, updated_at = ? WHERE id = ?
	`, status, time.Now(), messageID)
	return err
}

// GetMessages gets messages for a radio, optionally filtered
func (db *DB) GetMessages(radioID int, limit int, offset int) ([]*Message, error) {
	query := `
		SELECT id, radio_id, from_node_id, to_node_id, channel_id, content,
			   status, direction, port_num, timestamp, created_at, updated_at
		FROM messages
		WHERE radio_id = ?
		ORDER BY timestamp DESC
	`
	args := []interface{}{radioID}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.RadioID, &message.FromNodeID, &message.ToNodeID,
			&message.ChannelID, &message.Content, &message.Status, &message.Direction,
			&message.PortNum, &message.Timestamp, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

// GetRecentMessages gets recent messages for a radio
func (db *DB) GetRecentMessages(radioID int, limit int) ([]*Message, error) {
	return db.GetMessages(radioID, limit, 0)
}

// UpsertConfig creates or updates a config value
func (db *DB) UpsertConfig(radioID int, key, value, category string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		INSERT INTO configs (radio_id, key, value, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(radio_id, key) DO UPDATE SET
			value = excluded.value,
			category = excluded.category,
			updated_at = excluded.updated_at
	`, radioID, key, value, category, now, now)
	return err
}

// GetConfigs gets all config values for a radio
func (db *DB) GetConfigs(radioID int) (map[string]string, error) {
	rows, err := db.conn.Query(`
		SELECT key, value FROM configs WHERE radio_id = ? ORDER BY category, key
	`, radioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		configs[key] = value
	}
	return configs, rows.Err()
}
