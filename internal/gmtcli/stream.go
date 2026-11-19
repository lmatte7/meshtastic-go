package gmtcli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// broadcastNodeNumber is the destination of packets addressed to every node.
const broadcastNodeNumber = 0xffffffff

// streamConfigID tags our want_config_id requests; the radio echoes it back
// in config_complete_id when it has finished sending its database.
const streamConfigID = 7

// streamRefreshInterval is how often the radio is asked for its node and
// channel database again while streaming. Besides picking up renamed nodes
// and channels this acts as a keepalive: the firmware stops forwarding
// packets to a client it has not heard from in a while.
const streamRefreshInterval = 5 * time.Minute

// StreamChannel describes a channel configured on the radio.
type StreamChannel struct {
	Index uint32 `json:"index"`
	Name  string `json:"name"`
	Role  string `json:"role,omitempty"`
}

// StreamReady is emitted once, after the radio has sent its node and channel
// database, so a consumer knows its own node and the channel names before
// the first message arrives.
type StreamReady struct {
	Type       string          `json:"type"`
	ReceivedAt time.Time       `json:"received_at"`
	MyNode     *MeshNode       `json:"my_node,omitempty"`
	Channels   []StreamChannel `json:"channels"`
}

// StreamPacket is one received packet. Type is "message" for a text message
// and "packet" for anything else. For messages Text holds the body; for other
// decoded packets Payload holds the protobuf payload as JSON when the port is
// known, otherwise PayloadBase64 holds the raw bytes.
type StreamPacket struct {
	Type       string    `json:"type"`
	ReceivedAt time.Time `json:"received_at"`
	ID         uint32    `json:"id"`
	RxTime     string    `json:"rx_time,omitempty"`

	From      *MeshNode `json:"from"`
	To        *MeshNode `json:"to,omitempty"`
	Broadcast bool      `json:"broadcast"`
	// Direct is true when the packet was addressed to this radio.
	Direct bool `json:"direct"`
	// Outgoing is true when this radio sent the packet (for example from a
	// paired phone).
	Outgoing bool           `json:"outgoing"`
	Channel  *StreamChannel `json:"channel,omitempty"`

	Port          string          `json:"port,omitempty"`
	Text          string          `json:"text,omitempty"`
	Payload       json.RawMessage `json:"payload,omitempty"`
	PayloadBase64 string          `json:"payload_base64,omitempty"`
	Encrypted     bool            `json:"encrypted,omitempty"`
	ReplyID       uint32          `json:"reply_id,omitempty"`
	Emoji         uint32          `json:"emoji,omitempty"`
	RequestID     uint32          `json:"request_id,omitempty"`
	WantResponse  bool            `json:"want_response,omitempty"`

	SNR       float64 `json:"snr_db,omitempty"`
	RSSI      int32   `json:"rssi_dbm,omitempty"`
	HopLimit  uint32  `json:"hop_limit"`
	HopStart  uint32  `json:"hop_start,omitempty"`
	HopsTaken *uint32 `json:"hops_taken,omitempty"`
	ViaMQTT   bool    `json:"via_mqtt"`
	WantAck   bool    `json:"want_ack"`
	Priority  string  `json:"priority,omitempty"`
}

// portPayloadTypes maps application ports to the protobuf message carried in
// their payload, so --full output can show it as JSON.
var portPayloadTypes = map[gomeshproto.PortNum]func() proto.Message{
	gomeshproto.PortNum_POSITION_APP:        func() proto.Message { return &gomeshproto.Position{} },
	gomeshproto.PortNum_NODEINFO_APP:        func() proto.Message { return &gomeshproto.User{} },
	gomeshproto.PortNum_ROUTING_APP:         func() proto.Message { return &gomeshproto.Routing{} },
	gomeshproto.PortNum_ADMIN_APP:           func() proto.Message { return &gomeshproto.AdminMessage{} },
	gomeshproto.PortNum_WAYPOINT_APP:        func() proto.Message { return &gomeshproto.Waypoint{} },
	gomeshproto.PortNum_TELEMETRY_APP:       func() proto.Message { return &gomeshproto.Telemetry{} },
	gomeshproto.PortNum_TRACEROUTE_APP:      func() proto.Message { return &gomeshproto.RouteDiscovery{} },
	gomeshproto.PortNum_NEIGHBORINFO_APP:    func() proto.Message { return &gomeshproto.NeighborInfo{} },
	gomeshproto.PortNum_STORE_FORWARD_APP:   func() proto.Message { return &gomeshproto.StoreAndForward{} },
	gomeshproto.PortNum_PAXCOUNTER_APP:      func() proto.Message { return &gomeshproto.Paxcount{} },
	gomeshproto.PortNum_MAP_REPORT_APP:      func() proto.Message { return &gomeshproto.MapReport{} },
	gomeshproto.PortNum_REMOTE_HARDWARE_APP: func() proto.Message { return &gomeshproto.HardwareMessage{} },
}

// packetStream turns FromRadio packets into JSON lines, keeping a cache of
// the radio's node and channel database so every line carries sender and
// channel details.
type packetStream struct {
	radio    gomesh.Radio
	full     bool
	out      *json.Encoder
	now      func() time.Time
	myNode   uint32
	nodes    map[uint32]MeshNode
	channels map[uint32]StreamChannel
	ready    bool
}

func newPacketStream(radio gomesh.Radio, full bool, w io.Writer) *packetStream {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &packetStream{
		radio:    radio,
		full:     full,
		out:      enc,
		now:      time.Now,
		nodes:    make(map[uint32]MeshNode),
		channels: make(map[uint32]StreamChannel),
	}
}

// streamMessages is the action for the stream command.
func streamMessages(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := newPacketStream(radio, c.Bool("full"), os.Stdout)
	if err := s.run(ctx); err != nil {
		return cli.Exit(err.Error(), 1)
	}
	return nil
}

// run requests the radio's database, then forwards packets until the context
// is cancelled or the radio link fails.
func (s *packetStream) run(ctx context.Context) error {
	if err := s.requestConfig(); err != nil {
		return err
	}
	nextRefresh := s.now().Add(streamRefreshInterval)

	for ctx.Err() == nil {
		responses, err := s.radio.ReadResponse(false)
		if err != nil {
			return fmt.Errorf("reading from radio: %w", err)
		}
		for _, response := range responses {
			if err := s.handle(response); err != nil {
				return err
			}
		}
		if !s.now().Before(nextRefresh) {
			if err := s.requestConfig(); err != nil {
				return err
			}
			nextRefresh = s.now().Add(streamRefreshInterval)
		}
	}
	return nil
}

// requestConfig asks the radio to send its configuration, channel and node
// database. The answers are consumed by handle.
func (s *packetStream) requestConfig() error {
	request := gomeshproto.ToRadio{PayloadVariant: &gomeshproto.ToRadio_WantConfigId{WantConfigId: streamConfigID}}
	out, err := proto.Marshal(&request)
	if err != nil {
		return err
	}
	if err := s.radio.SendPacket(out); err != nil {
		return fmt.Errorf("requesting radio config: %w", err)
	}
	return nil
}

// handle processes one FromRadio message: database packets update the caches,
// mesh packets are written out according to the full setting.
func (s *packetStream) handle(response *gomeshproto.FromRadio) error {
	switch payload := response.GetPayloadVariant().(type) {
	case *gomeshproto.FromRadio_MyInfo:
		s.myNode = payload.MyInfo.GetMyNodeNum()
	case *gomeshproto.FromRadio_NodeInfo:
		node := ProtoNodeToMeshNode(payload)
		s.nodes[node.NodeNumber] = node
	case *gomeshproto.FromRadio_Channel:
		channel := payload.Channel
		if channel.GetRole() == gomeshproto.Channel_DISABLED {
			delete(s.channels, uint32(channel.GetIndex()))
			break
		}
		s.channels[uint32(channel.GetIndex())] = StreamChannel{
			Index: uint32(channel.GetIndex()),
			Name:  channel.GetSettings().GetName(),
			Role:  channel.GetRole().String(),
		}
	case *gomeshproto.FromRadio_ConfigCompleteId:
		if !s.ready {
			s.ready = true
			return s.out.Encode(s.readyEvent())
		}
	case *gomeshproto.FromRadio_Packet:
		return s.handlePacket(payload.Packet)
	}
	return nil
}

func (s *packetStream) readyEvent() StreamReady {
	ready := StreamReady{
		Type:       "ready",
		ReceivedAt: s.now().UTC(),
		MyNode:     s.node(s.myNode),
		Channels:   make([]StreamChannel, 0, len(s.channels)),
	}
	for i := uint32(0); i < 8; i++ {
		if channel, ok := s.channels[i]; ok {
			ready.Channels = append(ready.Channels, channel)
		}
	}
	return ready
}

// handlePacket writes a mesh packet as a line if it is a text message or if
// full output was requested, and keeps the node cache current from node info
// broadcasts.
func (s *packetStream) handlePacket(packet *gomeshproto.MeshPacket) error {
	if packet == nil {
		return nil
	}
	decoded := packet.GetDecoded()

	if decoded != nil && decoded.GetPortnum() == gomeshproto.PortNum_NODEINFO_APP {
		s.learnUser(packet.GetFrom(), decoded.GetPayload())
	}

	isText := decoded != nil && decoded.GetPortnum() == gomeshproto.PortNum_TEXT_MESSAGE_APP
	if !isText && !s.full {
		return nil
	}

	return s.out.Encode(s.packetEvent(packet))
}

// learnUser updates the node cache from a NODEINFO_APP broadcast so later
// lines from that node carry its name even if it was unknown at start-up.
func (s *packetStream) learnUser(from uint32, payload []byte) {
	user := &gomeshproto.User{}
	if err := proto.Unmarshal(payload, user); err != nil {
		return
	}
	node := s.nodes[from]
	node.NodeNumber = from
	user.Id = orDefault(user.Id, nodeIDFromNumber(from))
	node.ID = user.Id
	node.LongName = user.LongName
	node.ShortName = user.ShortName
	node.IsLicensed = user.IsLicensed
	node.Role = user.Role.String()
	if user.HwModel != gomeshproto.HardwareModel_UNSET {
		node.HardwareModel = user.HwModel.String()
	}
	if len(user.Macaddr) > 0 {
		node.MACAddress = net.HardwareAddr(user.Macaddr).String()
	}
	s.nodes[from] = node
}

// packetEvent builds the output line for a mesh packet.
func (s *packetStream) packetEvent(packet *gomeshproto.MeshPacket) StreamPacket {
	event := StreamPacket{
		Type:       "packet",
		ReceivedAt: s.now().UTC(),
		ID:         packet.GetId(),
		From:       s.node(packet.GetFrom()),
		Broadcast:  packet.GetTo() == broadcastNodeNumber,
		Direct:     packet.GetTo() == s.myNode,
		Outgoing:   packet.GetFrom() == s.myNode,
		SNR:        float64(packet.GetRxSnr()),
		RSSI:       packet.GetRxRssi(),
		HopLimit:   packet.GetHopLimit(),
		HopStart:   packet.GetHopStart(),
		ViaMQTT:    packet.GetViaMqtt(),
		WantAck:    packet.GetWantAck(),
		Priority:   packet.GetPriority().String(),
	}
	if !event.Broadcast {
		event.To = s.node(packet.GetTo())
	}
	if packet.GetRxTime() != 0 {
		event.RxTime = time.Unix(int64(packet.GetRxTime()), 0).UTC().Format(time.RFC3339)
	}
	if packet.GetHopStart() > 0 && packet.GetHopStart() >= packet.GetHopLimit() {
		hops := packet.GetHopStart() - packet.GetHopLimit()
		event.HopsTaken = &hops
	}
	decoded := packet.GetDecoded()
	if decoded == nil {
		// The radio could not decrypt this packet, so the channel field is
		// the channel hash rather than an index and the payload is ciphertext.
		event.Encrypted = true
		event.Channel = &StreamChannel{Index: packet.GetChannel()}
		event.PayloadBase64 = base64.StdEncoding.EncodeToString(packet.GetEncrypted())
		return event
	}

	event.Channel = s.channel(packet.GetChannel())
	event.Port = decoded.GetPortnum().String()
	event.ReplyID = decoded.GetReplyId()
	event.Emoji = decoded.GetEmoji()
	event.RequestID = decoded.GetRequestId()
	event.WantResponse = decoded.GetWantResponse()

	if decoded.GetPortnum() == gomeshproto.PortNum_TEXT_MESSAGE_APP {
		event.Type = "message"
		event.Text = string(decoded.GetPayload())
		return event
	}

	if payload, ok := decodePayload(decoded.GetPortnum(), decoded.GetPayload()); ok {
		event.Payload = payload
	} else {
		event.PayloadBase64 = base64.StdEncoding.EncodeToString(decoded.GetPayload())
	}
	return event
}

// decodePayload renders a known application payload as JSON.
func decodePayload(port gomeshproto.PortNum, payload []byte) (json.RawMessage, bool) {
	newMessage, ok := portPayloadTypes[port]
	if !ok {
		return nil, false
	}
	message := newMessage()
	if err := proto.Unmarshal(payload, message); err != nil {
		return nil, false
	}
	out, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(message)
	if err != nil {
		return nil, false
	}
	return json.RawMessage(out), true
}

// node returns the cached details for a node, or the little that can be
// derived from its number when the radio has not told us about it.
func (s *packetStream) node(number uint32) *MeshNode {
	if node, ok := s.nodes[number]; ok {
		return &node
	}
	return &MeshNode{NodeNumber: number, ID: nodeIDFromNumber(number)}
}

// channel returns the cached details for a channel index.
func (s *packetStream) channel(index uint32) *StreamChannel {
	if channel, ok := s.channels[index]; ok {
		return &channel
	}
	return &StreamChannel{Index: index}
}

// nodeIDFromNumber derives the "!xxxxxxxx" node ID Meshtastic uses from a
// node number.
func nodeIDFromNumber(number uint32) string {
	return fmt.Sprintf("!%08x", number)
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
