package gmtcli

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lmatte7/meshtastic-go/internal/gomesh"
	"github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// broadcastNodeNumber is the destination of packets addressed to every node.
const broadcastNodeNumber = 0xffffffff

// streamConfigID tags our want_config_id requests; the radio echoes it back
// in config_complete_id when it has finished sending its database.
const streamConfigID = 7

// ackTimeout is how long a sent message waits for the radio to report its
// fate before it is reported as failed.
const ackTimeout = 5 * time.Minute

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
// known, otherwise PayloadBase64 holds the raw bytes. ReceivedAt is when the
// stream received the packet; the radio's own timestamp is not reported
// because radios without a time source stamp with a clock that can be days
// off.
type StreamPacket struct {
	Type       string    `json:"type"`
	ReceivedAt time.Time `json:"received_at"`
	ID         uint32    `json:"id"`

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
	Reply         *StreamReply    `json:"reply,omitempty"`
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

// StreamCommand is one line read from stdin with --commands: a text message
// to send. To may be a node number, a "!xxxxxxxx" node id, or a node's long
// or short name (which must match exactly one known node); leave it out to
// broadcast on the channel. ID is echoed back in the resulting event.
type StreamCommand struct {
	ID string `json:"id,omitempty"`
	// Command selects what to do. Empty or "send" sends Text; "nodes" asks
	// for the cached node list, answered with a StreamNodes event.
	Command string          `json:"command,omitempty"`
	To      json.RawMessage `json:"to,omitempty"`
	Channel uint32          `json:"channel"`
	Text    string          `json:"text"`
}

// StreamNodes answers a "nodes" command with every node the radio knows
// about, as the stream currently has them cached.
type StreamNodes struct {
	Type   string     `json:"type"`
	At     time.Time  `json:"at"`
	ID     string     `json:"id,omitempty"`
	MyNode *MeshNode  `json:"my_node,omitempty"`
	Nodes  []MeshNode `json:"nodes"`
}

// StreamReply is attached to every message so a bot can answer without
// assembling addresses itself: copy Direct to message the sender privately
// or Channel to answer publicly where the message arrived, add "text", and
// write it to stdin with --commands.
type StreamReply struct {
	Direct  StreamCommandTemplate `json:"direct"`
	Channel StreamCommandTemplate `json:"channel"`
}

// StreamCommandTemplate is a StreamCommand without the text.
type StreamCommandTemplate struct {
	To      string `json:"to,omitempty"`
	Channel uint32 `json:"channel"`
}

// resolveRecipient turns the "to" of a command into a node number. It
// returns 0 for a broadcast.
func (s *packetStream) resolveRecipient(raw json.RawMessage) (uint32, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}

	var number uint32
	if err := json.Unmarshal(raw, &number); err == nil {
		return number, nil
	}

	var name string
	if err := json.Unmarshal(raw, &name); err != nil {
		return 0, fmt.Errorf("to must be a node number, a \"!xxxxxxxx\" node id or a node name")
	}
	if strings.TrimSpace(name) == "" {
		return 0, nil
	}
	return lookupNode(name, s.nodes)
}

// lookupNode resolves a node given as a decimal node number, a "!xxxxxxxx"
// node id, or a long or short name that matches exactly one known node.
func lookupNode(ref string, nodes map[uint32]MeshNode) (uint32, error) {
	ref = strings.TrimSpace(ref)
	if number, err := strconv.ParseUint(ref, 10, 32); err == nil {
		return uint32(number), nil
	}
	if strings.HasPrefix(ref, "!") {
		number, err := strconv.ParseUint(ref[1:], 16, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid node id %q", ref)
		}
		return uint32(number), nil
	}

	var matches []uint32
	for num, node := range nodes {
		if strings.EqualFold(node.LongName, ref) || strings.EqualFold(node.ShortName, ref) {
			matches = append(matches, num)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return 0, fmt.Errorf("no known node named %q", ref)
	default:
		return 0, fmt.Errorf("%d nodes are named %q; use the node id instead", len(matches), ref)
	}
}

// StreamSent reports the outcome of a StreamCommand. Type is "sent" when the
// message was handed to the radio and "error" otherwise. PacketID is the
// mesh packet id, which later StreamDelivery events refer to.
type StreamSent struct {
	Type      string         `json:"type"`
	At        time.Time      `json:"at"`
	ID        string         `json:"id,omitempty"`
	PacketID  uint32         `json:"packet_id,omitempty"`
	To        *MeshNode      `json:"to,omitempty"`
	Broadcast bool           `json:"broadcast"`
	Channel   *StreamChannel `json:"channel,omitempty"`
	Text      string         `json:"text,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// StreamDelivery reports what became of a message sent with --commands.
// Type is "queued" once the radio has accepted the packet for transmission,
// "delivered" when it was acknowledged, and "failed" otherwise. For
// broadcasts, and whenever the acknowledgement comes from a node other than
// the recipient, Implicit is true: a neighbour rebroadcast the packet, which
// proves it left this radio but not that the recipient got it.
type StreamDelivery struct {
	Type      string         `json:"type"`
	At        time.Time      `json:"at"`
	ID        string         `json:"id,omitempty"`
	PacketID  uint32         `json:"packet_id"`
	To        *MeshNode      `json:"to,omitempty"`
	Broadcast bool           `json:"broadcast"`
	Channel   *StreamChannel `json:"channel,omitempty"`
	// From is the node that sent the acknowledgement or error.
	From     *MeshNode `json:"from,omitempty"`
	Implicit bool      `json:"implicit,omitempty"`
	Error    string    `json:"error,omitempty"`
}

// pendingSend is a message we sent and have not heard back about yet.
type pendingSend struct {
	id      string
	to      uint32 // 0 for a broadcast
	channel uint32
	sentAt  time.Time
}

// packetStream turns FromRadio packets into JSON lines, keeping a cache of
// the radio's node and channel database so every line carries sender and
// channel details. With commands set it also sends text messages read from
// that channel.
type packetStream struct {
	radio    gomesh.Radio
	full     bool
	out      *json.Encoder
	now      func() time.Time
	send     func(text string, to, channel uint32) (packetID uint32, err error)
	commands <-chan string
	myNode   uint32
	nodes    map[uint32]MeshNode
	channels map[uint32]StreamChannel
	pending  map[uint32]pendingSend
	ready    bool

	// lastPacketAt is when the stream itself last received a packet from
	// each node. It is the only last-heard time worth reporting: the stream
	// sees packets as they arrive, while a radio without a time source
	// stamps them with a clock that can be days off.
	lastPacketAt map[uint32]time.Time
}

func newPacketStream(radio gomesh.Radio, full bool, w io.Writer) *packetStream {
	// One compact JSON document per line; pipe through jq for a readable view.
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	s := &packetStream{
		radio:    radio,
		full:     full,
		out:      enc,
		now:      time.Now,
		nodes:    make(map[uint32]MeshNode),
		channels: make(map[uint32]StreamChannel),
		pending:  make(map[uint32]pendingSend),

		lastPacketAt: make(map[uint32]time.Time),
	}
	s.send = s.sendText
	return s
}

// withStreamAge returns a copy of a cached node whose last-heard fields say
// when the stream itself last received a packet from it. Nodes the stream has
// not heard during this run carry no last-heard fields: the radio's own
// stamps are not used, because a radio without a time source stamps with a
// clock that can be days off.
func (s *packetStream) withStreamAge(node MeshNode) MeshNode {
	node.LastHeard, node.LastHeardSecondsAgo, node.LastHeardHuman = "", 0, ""
	lastSeen, ok := s.lastPacketAt[node.NodeNumber]
	if !ok {
		return node
	}
	since := s.now().Sub(lastSeen)
	node.LastHeard = lastSeen.UTC().Format(time.RFC3339)
	node.LastHeardSecondsAgo = int(since.Seconds())
	node.LastHeardHuman = since.Round(time.Second).String()
	return node
}

// sendText sends a text message and returns the packet id the library
// assigned, so the radio's later queue status and acknowledgement can be
// matched back to the command. A zero "to" broadcasts on the channel.
func (s *packetStream) sendText(text string, to, channel uint32) (uint32, error) {
	if len(text) > int(gomeshproto.Constants_DATA_PAYLOAD_LEN) {
		return 0, fmt.Errorf("message too large: %d bytes, maximum %d", len(text), gomeshproto.Constants_DATA_PAYLOAD_LEN)
	}
	return s.radio.SendTextMessageWithID(text, int64(to), int64(channel))
}

// streamMessages is the action for the stream command.
func streamMessages(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := newPacketStream(radio, c.Bool("full"), os.Stdout)
	if c.Bool("commands") {
		s.commands = readLines(os.Stdin)
	}
	if err := s.run(ctx); err != nil {
		return cli.Exit(err.Error(), 1)
	}
	return nil
}

// readLines delivers each line of r on the returned channel and closes it at
// end of input.
func readLines(r io.Reader) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()
	return lines
}

// run requests the radio's database, then forwards packets (and sends any
// queued commands) until the context is cancelled, stdin is closed in
// --commands mode, or the radio link fails.
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
		open, err := s.drainCommands()
		if err != nil {
			return err
		}
		if !open {
			// The parent closed our stdin, which is its way of asking us to
			// stop.
			return nil
		}
		if err := s.expirePending(); err != nil {
			return err
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

// drainCommands sends every command line waiting on stdin. It reports false
// once stdin has been closed and no more lines will arrive.
func (s *packetStream) drainCommands() (open bool, err error) {
	if s.commands == nil {
		return true, nil
	}
	for {
		select {
		case line, ok := <-s.commands:
			if !ok {
				return false, nil
			}
			if err := s.execute(line); err != nil {
				return true, err
			}
		default:
			return true, nil
		}
	}
}

// execute parses one command line, sends the message and writes the outcome.
// Problems with the command itself are reported as an "error" event rather
// than returned, so one bad line does not take the stream down.
func (s *packetStream) execute(line string) error {
	if strings.TrimSpace(line) == "" {
		return nil
	}

	var command StreamCommand
	if err := json.Unmarshal([]byte(line), &command); err != nil {
		return s.out.Encode(StreamSent{Type: "error", At: s.now().UTC(), Error: "invalid command: " + err.Error()})
	}

	switch command.Command {
	case "", "send":
	case "nodes":
		return s.out.Encode(s.nodesEvent(command.ID))
	default:
		return s.out.Encode(StreamSent{Type: "error", At: s.now().UTC(), ID: command.ID, Error: fmt.Sprintf("unknown command %q", command.Command)})
	}

	result := StreamSent{
		Type:    "sent",
		At:      s.now().UTC(),
		ID:      command.ID,
		Channel: s.channel(command.Channel),
		Text:    command.Text,
	}

	to, err := s.resolveRecipient(command.To)
	result.Broadcast = to == 0 || to == broadcastNodeNumber
	if !result.Broadcast {
		result.To = s.node(to)
	}

	switch {
	case err != nil:
		result.Type, result.Error = "error", err.Error()
	case command.Text == "":
		result.Type, result.Error = "error", "text is required"
	case command.Channel > 7:
		result.Type, result.Error = "error", "channel must be between 0 and 7"
	default:
		packetID, err := s.send(command.Text, to, command.Channel)
		if err != nil {
			result.Type, result.Error = "error", err.Error()
			break
		}
		result.PacketID = packetID
		if result.Broadcast {
			to = 0
		}
		s.pending[packetID] = pendingSend{id: command.ID, to: to, channel: command.Channel, sentAt: s.now()}
	}
	return s.out.Encode(result)
}

// nodesEvent builds the answer to a "nodes" command: the cached node list in
// node number order.
func (s *packetStream) nodesEvent(id string) StreamNodes {
	event := StreamNodes{
		Type:   "nodes",
		At:     s.now().UTC(),
		ID:     id,
		MyNode: s.node(s.myNode),
		Nodes:  make([]MeshNode, 0, len(s.nodes)),
	}
	for _, node := range s.nodes {
		event.Nodes = append(event.Nodes, s.withStreamAge(node))
	}
	sort.Slice(event.Nodes, func(i, j int) bool { return event.Nodes[i].NodeNumber < event.Nodes[j].NodeNumber })
	return event
}

// deliveryEvent builds a StreamDelivery for a pending message.
func (s *packetStream) deliveryEvent(eventType string, packetID uint32, p pendingSend) StreamDelivery {
	event := StreamDelivery{
		Type:      eventType,
		At:        s.now().UTC(),
		ID:        p.id,
		PacketID:  packetID,
		Broadcast: p.to == 0,
		Channel:   s.channel(p.channel),
	}
	if !event.Broadcast {
		event.To = s.node(p.to)
	}
	return event
}

// handleQueueStatus reports a pending message as queued once the radio has
// taken it, or as failed if the radio's queue rejected it.
func (s *packetStream) handleQueueStatus(status *gomeshproto.QueueStatus) error {
	p, ok := s.pending[status.GetMeshPacketId()]
	if !ok {
		return nil
	}
	event := s.deliveryEvent("queued", status.GetMeshPacketId(), p)
	if status.GetRes() != 0 {
		event.Type = "failed"
		event.Error = fmt.Sprintf("radio queue error %d", status.GetRes())
		delete(s.pending, status.GetMeshPacketId())
	}
	return s.out.Encode(event)
}

// handleAck matches a routing packet to a pending message and reports it as
// delivered or failed. Routing packets for other requests are ignored.
func (s *packetStream) handleAck(packet *gomeshproto.MeshPacket, decoded *gomeshproto.Data) error {
	p, ok := s.pending[decoded.GetRequestId()]
	if !ok {
		return nil
	}
	delete(s.pending, decoded.GetRequestId())

	event := s.deliveryEvent("delivered", decoded.GetRequestId(), p)
	event.From = s.node(packet.GetFrom())
	event.Implicit = p.to == 0 || packet.GetFrom() != p.to

	routing := &gomeshproto.Routing{}
	if err := proto.Unmarshal(decoded.GetPayload(), routing); err == nil {
		if reason := routing.GetErrorReason(); reason != gomeshproto.Routing_NONE {
			event.Type = "failed"
			event.Error = reason.String()
			event.Implicit = false
		}
	}
	return s.out.Encode(event)
}

// expirePending reports messages the radio never answered for as failed.
func (s *packetStream) expirePending() error {
	for packetID, p := range s.pending {
		if s.now().Sub(p.sentAt) < ackTimeout {
			continue
		}
		delete(s.pending, packetID)
		event := s.deliveryEvent("failed", packetID, p)
		event.Error = fmt.Sprintf("no acknowledgement within %s", ackTimeout)
		if err := s.out.Encode(event); err != nil {
			return err
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
	case *gomeshproto.FromRadio_QueueStatus:
		return s.handleQueueStatus(payload.QueueStatus)
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

	// Every received packet, printed or not, is timed: its receipt is when
	// the sending node was last heard.
	s.lastPacketAt[packet.GetFrom()] = s.now()

	if decoded != nil && decoded.GetPortnum() == gomeshproto.PortNum_NODEINFO_APP {
		s.learnUser(packet.GetFrom(), decoded.GetPayload())
	}
	if decoded != nil && decoded.GetPortnum() == gomeshproto.PortNum_ROUTING_APP {
		if err := s.handleAck(packet, decoded); err != nil {
			return err
		}
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
	if mac := macAddress(user); mac != "" {
		node.MACAddress = mac
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
		event.Reply = &StreamReply{
			Direct:  StreamCommandTemplate{To: event.From.ID, Channel: packet.GetChannel()},
			Channel: StreamCommandTemplate{Channel: packet.GetChannel()},
		}
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
	return roundPayloadNumbers(out), true
}

// payloadDecimals is how many decimal places known protobuf fields keep when
// a decoded payload is printed. Anything not listed keeps defaultDecimals.
var payloadDecimals = map[string]int{
	"air_util_tx":          2,
	"channel_utilization":  2,
	"battery_level":        1,
	"voltage":              3,
	"current":              3,
	"temperature":          2,
	"relative_humidity":    2,
	"barometric_pressure":  2,
	"gas_resistance":       1,
	"lux":                  2,
	"white_lux":            2,
	"ir_lux":               2,
	"uv_lux":               2,
	"wind_speed":           2,
	"wind_gust":            2,
	"wind_lull":            2,
	"rainfall_1h":          2,
	"rainfall_24h":         2,
	"soil_temperature":     2,
	"iaq":                  1,
	"distance":             2,
	"weight":               3,
	"radiation":            3,
	"ch1_voltage":          3,
	"ch2_voltage":          3,
	"ch3_voltage":          3,
	"ch1_current":          3,
	"ch2_current":          3,
	"ch3_current":          3,
	"snr":                  3,
	"rx_snr":               3,
	"ground_speed":         2,
	"ground_track":         2,
	"gps_accuracy":         2,
	"altitude_hae":         2,
	"altitude_geoidal_sep": 2,
}

const defaultDecimals = 4

// roundPayloadNumbers tidies a decoded payload's JSON for people and
// programs alike: numbers are rounded to the precision they actually carry
// (float32 readings otherwise print with every digit a float64 can hold),
// integers are left alone, and a macaddr, which protobuf JSON renders as
// base64, becomes the usual colon-separated hex. Field order is not
// preserved (keys come out sorted), which is fine for a payload blob.
func roundPayloadNumbers(raw []byte) json.RawMessage {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return raw
	}
	out, err := json.Marshal(roundValue(value, ""))
	if err != nil {
		return raw
	}
	return out
}

func roundValue(v any, key string) any {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			x[k] = roundValue(child, k)
		}
		return x
	case []any:
		for i, child := range x {
			x[i] = roundValue(child, key)
		}
		return x
	case string:
		if key == "macaddr" {
			if mac, err := base64.StdEncoding.DecodeString(x); err == nil && len(mac) > 0 {
				return net.HardwareAddr(mac).String()
			}
		}
		return x
	case json.Number:
		if !strings.ContainsAny(x.String(), ".eE") {
			return x // an integer
		}
		f, err := x.Float64()
		if err != nil {
			return x
		}
		places, ok := payloadDecimals[key]
		if !ok {
			places = defaultDecimals
		}
		return roundTo(f, places)
	default:
		return v
	}
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
