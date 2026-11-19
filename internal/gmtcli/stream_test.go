package gmtcli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"google.golang.org/protobuf/proto"
)

const (
	testMyNode    = uint32(1397525059)
	testPeerNode  = uint32(1135973780)
	testOtherNode = uint32(994488212)
)

// newTestStream returns a stream with a fixed clock that has been fed the
// radio's database: our own node, one known peer and two channels.
func newTestStream(t *testing.T, full bool) (*packetStream, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	s := newPacketStream(gomesh.Radio{}, full, &buf)
	s.now = func() time.Time { return time.Date(2026, 11, 19, 4, 12, 8, 0, time.UTC) }

	feed := []*gomeshproto.FromRadio{
		{PayloadVariant: &gomeshproto.FromRadio_MyInfo{MyInfo: &gomeshproto.MyNodeInfo{MyNodeNum: testMyNode}}},
		{PayloadVariant: &gomeshproto.FromRadio_NodeInfo{NodeInfo: &gomeshproto.NodeInfo{
			Num:  testMyNode,
			User: &gomeshproto.User{Id: "!534c8a43", LongName: "Meshtastic 8a43", ShortName: "8a43", HwModel: gomeshproto.HardwareModel_T_ECHO},
		}}},
		{PayloadVariant: &gomeshproto.FromRadio_NodeInfo{NodeInfo: &gomeshproto.NodeInfo{
			Num:  testPeerNode,
			User: &gomeshproto.User{Id: "!43b59594", LongName: "DvdGmz01", ShortName: "DG01", HwModel: gomeshproto.HardwareModel_HELTEC_V3},
		}}},
		{PayloadVariant: &gomeshproto.FromRadio_Channel{Channel: &gomeshproto.Channel{
			Index: 0, Role: gomeshproto.Channel_PRIMARY, Settings: &gomeshproto.ChannelSettings{Name: ""},
		}}},
		{PayloadVariant: &gomeshproto.FromRadio_Channel{Channel: &gomeshproto.Channel{
			Index: 1, Role: gomeshproto.Channel_SECONDARY, Settings: &gomeshproto.ChannelSettings{Name: "bots"},
		}}},
		{PayloadVariant: &gomeshproto.FromRadio_Channel{Channel: &gomeshproto.Channel{Index: 2, Role: gomeshproto.Channel_DISABLED}}},
		{PayloadVariant: &gomeshproto.FromRadio_ConfigCompleteId{ConfigCompleteId: streamConfigID}},
	}
	for _, response := range feed {
		if err := s.handle(response); err != nil {
			t.Fatalf("handle(%v): %v", response, err)
		}
	}
	return s, &buf
}

func textPacket(from, to uint32, channel uint32, text string) *gomeshproto.FromRadio {
	return &gomeshproto.FromRadio{PayloadVariant: &gomeshproto.FromRadio_Packet{Packet: &gomeshproto.MeshPacket{
		From: from, To: to, Channel: channel, Id: 1836540931, RxTime: 1795061528,
		RxSnr: -1.75, RxRssi: -110, HopLimit: 2, HopStart: 3,
		PayloadVariant: &gomeshproto.MeshPacket_Decoded{Decoded: &gomeshproto.Data{
			Portnum: gomeshproto.PortNum_TEXT_MESSAGE_APP, Payload: []byte(text),
		}},
	}}}
}

func lines(buf *bytes.Buffer) []map[string]interface{} {
	var out []map[string]interface{}
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			panic("invalid JSON line: " + line + ": " + err.Error())
		}
		out = append(out, m)
	}
	return out
}

func TestStreamReadyLine(t *testing.T) {
	_, buf := newTestStream(t, false)
	got := lines(buf)
	if len(got) != 1 || got[0]["type"] != "ready" {
		t.Fatalf("expected a single ready line, got %v", got)
	}
	myNode := got[0]["my_node"].(map[string]interface{})
	if myNode["node_id"] != "!534c8a43" {
		t.Errorf("my_node = %v", myNode)
	}
	channels := got[0]["channels"].([]interface{})
	if len(channels) != 2 {
		t.Fatalf("expected the two enabled channels, got %v", channels)
	}
	if channels[1].(map[string]interface{})["name"] != "bots" {
		t.Errorf("channels = %v", channels)
	}
}

func TestStreamDirectMessage(t *testing.T) {
	s, buf := newTestStream(t, false)
	buf.Reset()

	if err := s.handle(textPacket(testPeerNode, testMyNode, 0, "hi <there>")); err != nil {
		t.Fatal(err)
	}
	got := lines(buf)
	if len(got) != 1 {
		t.Fatalf("expected one line, got %d", len(got))
	}
	m := got[0]
	if m["type"] != "message" || m["text"] != "hi <there>" {
		t.Errorf("type/text = %v/%v", m["type"], m["text"])
	}
	if m["direct"] != true || m["broadcast"] != false || m["outgoing"] != false {
		t.Errorf("direct/broadcast/outgoing = %v/%v/%v", m["direct"], m["broadcast"], m["outgoing"])
	}
	from := m["from"].(map[string]interface{})
	if from["longname"] != "DvdGmz01" || from["node_number"] != float64(testPeerNode) {
		t.Errorf("from = %v", from)
	}
	to := m["to"].(map[string]interface{})
	if to["node_id"] != "!534c8a43" {
		t.Errorf("to = %v", to)
	}
	if m["rx_time"] != "2026-11-19T04:12:08Z" || m["received_at"] != "2026-11-19T04:12:08Z" {
		t.Errorf("rx_time/received_at = %v/%v", m["rx_time"], m["received_at"])
	}
	if m["hops_taken"] != float64(1) || m["snr_db"] != -1.75 || m["rssi_dbm"] != float64(-110) {
		t.Errorf("radio fields = hops %v snr %v rssi %v", m["hops_taken"], m["snr_db"], m["rssi_dbm"])
	}
	if !strings.Contains(buf.String(), `"text":"hi <there>"`) {
		t.Errorf("text should be written without HTML escaping, got %s", buf.String())
	}
}

func TestStreamChannelMessageFromUnknownNode(t *testing.T) {
	s, buf := newTestStream(t, false)
	buf.Reset()

	if err := s.handle(textPacket(testOtherNode, broadcastNodeNumber, 1, "hello bots")); err != nil {
		t.Fatal(err)
	}
	m := lines(buf)[0]
	if m["broadcast"] != true || m["direct"] != false {
		t.Errorf("broadcast/direct = %v/%v", m["broadcast"], m["direct"])
	}
	if _, hasTo := m["to"]; hasTo {
		t.Error("broadcast messages should not carry a recipient")
	}
	channel := m["channel"].(map[string]interface{})
	if channel["index"] != float64(1) || channel["name"] != "bots" || channel["role"] != "SECONDARY" {
		t.Errorf("channel = %v", channel)
	}
	from := m["from"].(map[string]interface{})
	if from["node_id"] != "!3b46af94" || from["node_number"] != float64(testOtherNode) {
		t.Errorf("unknown sender should still get an id and number, got %v", from)
	}
}

func TestStreamSkipsNonTextWithoutFull(t *testing.T) {
	s, buf := newTestStream(t, false)
	buf.Reset()

	position, _ := proto.Marshal(&gomeshproto.Position{LatitudeI: 361824256, LongitudeI: -1151926272, Altitude: 471})
	responses := []*gomeshproto.FromRadio{
		{PayloadVariant: &gomeshproto.FromRadio_Packet{Packet: &gomeshproto.MeshPacket{
			From: testPeerNode, To: broadcastNodeNumber, Id: 7,
			PayloadVariant: &gomeshproto.MeshPacket_Decoded{Decoded: &gomeshproto.Data{Portnum: gomeshproto.PortNum_POSITION_APP, Payload: position}},
		}}},
		{PayloadVariant: &gomeshproto.FromRadio_Packet{Packet: &gomeshproto.MeshPacket{
			From: testOtherNode, To: broadcastNodeNumber, Id: 8, Channel: 0xab,
			PayloadVariant: &gomeshproto.MeshPacket_Encrypted{Encrypted: []byte{1, 2, 3}},
		}}},
	}
	for _, r := range responses {
		if err := s.handle(r); err != nil {
			t.Fatal(err)
		}
	}
	if buf.Len() != 0 {
		t.Fatalf("non-text packets should be dropped without --full, got %s", buf.String())
	}

	// The same packets are reported with --full.
	full, fullBuf := newTestStream(t, true)
	fullBuf.Reset()
	for _, r := range responses {
		if err := full.handle(r); err != nil {
			t.Fatal(err)
		}
	}
	got := lines(fullBuf)
	if len(got) != 2 {
		t.Fatalf("expected two lines with --full, got %d", len(got))
	}
	pos := got[0]
	if pos["type"] != "packet" || pos["port"] != "POSITION_APP" {
		t.Errorf("position line = %v", pos)
	}
	payload := pos["payload"].(map[string]interface{})
	if payload["latitude_i"] != float64(361824256) || payload["altitude"] != float64(471) {
		t.Errorf("decoded position payload = %v", payload)
	}
	enc := got[1]
	if enc["encrypted"] != true || enc["payload_base64"] != "AQID" || enc["port"] != nil {
		t.Errorf("encrypted line = %v", enc)
	}
	if enc["channel"].(map[string]interface{})["index"] != float64(0xab) {
		t.Errorf("encrypted packets should report the channel hash, got %v", enc["channel"])
	}
}

func TestStreamLearnsNodesFromNodeInfoBroadcasts(t *testing.T) {
	s, buf := newTestStream(t, false)
	buf.Reset()

	user, _ := proto.Marshal(&gomeshproto.User{Id: "!3b46af94", LongName: "WS7R G2 Base", ShortName: "W7G2", HwModel: gomeshproto.HardwareModel_STATION_G2})
	nodeInfo := &gomeshproto.FromRadio{PayloadVariant: &gomeshproto.FromRadio_Packet{Packet: &gomeshproto.MeshPacket{
		From: testOtherNode, To: broadcastNodeNumber,
		PayloadVariant: &gomeshproto.MeshPacket_Decoded{Decoded: &gomeshproto.Data{Portnum: gomeshproto.PortNum_NODEINFO_APP, Payload: user}},
	}}}
	if err := s.handle(nodeInfo); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("node info should not be printed without --full, got %s", buf.String())
	}
	if err := s.handle(textPacket(testOtherNode, testMyNode, 0, "now you know me")); err != nil {
		t.Fatal(err)
	}
	from := lines(buf)[0]["from"].(map[string]interface{})
	if from["longname"] != "WS7R G2 Base" || from["hardware_model"] != "STATION_G2" {
		t.Errorf("sender details should come from the node info broadcast, got %v", from)
	}
}

func TestStreamReplyTemplates(t *testing.T) {
	s, buf := newTestStream(t, false)
	buf.Reset()
	if err := s.handle(textPacket(testOtherNode, broadcastNodeNumber, 1, "!weather")); err != nil {
		t.Fatal(err)
	}
	reply := lines(buf)[0]["reply"].(map[string]interface{})
	direct := reply["direct"].(map[string]interface{})
	channel := reply["channel"].(map[string]interface{})
	if direct["to"] != "!3b46af94" || direct["channel"] != float64(1) {
		t.Errorf("direct reply template = %v", direct)
	}
	if _, hasTo := channel["to"]; hasTo || channel["channel"] != float64(1) {
		t.Errorf("channel reply template = %v", channel)
	}
}

// fakeSender records the messages a stream asked the radio to send and hands
// out sequential packet ids starting at 1001.
type fakeSender struct {
	calls  []string
	err    error
	nextID uint32
}

func (f *fakeSender) send(text string, to, channel uint32) (uint32, error) {
	f.calls = append(f.calls, fmt.Sprintf("%s|%d|%d", text, to, channel))
	if f.err != nil {
		return 0, f.err
	}
	if f.nextID == 0 {
		f.nextID = 1000
	}
	f.nextID++
	return f.nextID, nil
}

// routingPacket is an acknowledgement or error for requestID from a node.
func routingPacket(from, requestID uint32, reason gomeshproto.Routing_Error) *gomeshproto.FromRadio {
	payload, _ := proto.Marshal(&gomeshproto.Routing{Variant: &gomeshproto.Routing_ErrorReason{ErrorReason: reason}})
	return &gomeshproto.FromRadio{PayloadVariant: &gomeshproto.FromRadio_Packet{Packet: &gomeshproto.MeshPacket{
		From: from, To: testMyNode, Id: 99,
		PayloadVariant: &gomeshproto.MeshPacket_Decoded{Decoded: &gomeshproto.Data{
			Portnum: gomeshproto.PortNum_ROUTING_APP, RequestId: requestID, Payload: payload,
		}},
	}}}
}

func TestStreamDeliveryTracking(t *testing.T) {
	s, buf := newTestStream(t, false)
	sender := &fakeSender{}
	s.send = sender.send

	// A DM: sent -> queued -> delivered by the recipient.
	buf.Reset()
	_ = s.execute(`{"id":"dm-1","to":"DG01","text":"ping"}`)
	sent := lines(buf)[0]
	if sent["packet_id"] != float64(1001) {
		t.Fatalf("sent event should carry the packet id, got %v", sent)
	}

	buf.Reset()
	queue := &gomeshproto.FromRadio{PayloadVariant: &gomeshproto.FromRadio_QueueStatus{QueueStatus: &gomeshproto.QueueStatus{MeshPacketId: 1001}}}
	if err := s.handle(queue); err != nil {
		t.Fatal(err)
	}
	queued := lines(buf)[0]
	if queued["type"] != "queued" || queued["id"] != "dm-1" || queued["packet_id"] != float64(1001) {
		t.Errorf("queued event = %v", queued)
	}

	buf.Reset()
	if err := s.handle(routingPacket(testPeerNode, 1001, gomeshproto.Routing_NONE)); err != nil {
		t.Fatal(err)
	}
	got := lines(buf)
	if len(got) != 1 {
		t.Fatalf("expected only the delivery event (routing packets are not messages), got %v", got)
	}
	delivered := got[0]
	if delivered["type"] != "delivered" || delivered["id"] != "dm-1" || delivered["implicit"] != nil {
		t.Errorf("delivered event = %v", delivered)
	}
	if delivered["from"].(map[string]interface{})["shortname"] != "DG01" || delivered["to"].(map[string]interface{})["shortname"] != "DG01" {
		t.Errorf("delivered from/to = %v / %v", delivered["from"], delivered["to"])
	}
	if _, still := s.pending[1001]; still {
		t.Error("acknowledged packet should no longer be pending")
	}

	// A broadcast acknowledged by a neighbour is only an implicit ack.
	buf.Reset()
	_ = s.execute(`{"id":"bc-1","channel":1,"text":"hello all"}`)
	buf.Reset()
	if err := s.handle(routingPacket(testOtherNode, 1002, gomeshproto.Routing_NONE)); err != nil {
		t.Fatal(err)
	}
	implicit := lines(buf)[0]
	if implicit["type"] != "delivered" || implicit["implicit"] != true || implicit["broadcast"] != true || implicit["id"] != "bc-1" {
		t.Errorf("implicit ack event = %v", implicit)
	}
	if implicit["channel"].(map[string]interface{})["name"] != "bots" {
		t.Errorf("implicit ack channel = %v", implicit["channel"])
	}

	// A routing error marks the message failed.
	buf.Reset()
	_ = s.execute(`{"id":"dm-2","to":"DG01","text":"anyone there?"}`)
	buf.Reset()
	if err := s.handle(routingPacket(testMyNode, 1003, gomeshproto.Routing_MAX_RETRANSMIT)); err != nil {
		t.Fatal(err)
	}
	failed := lines(buf)[0]
	if failed["type"] != "failed" || failed["error"] != "MAX_RETRANSMIT" || failed["id"] != "dm-2" {
		t.Errorf("failed event = %v", failed)
	}

	// Routing packets for requests we did not make are ignored.
	buf.Reset()
	if err := s.handle(routingPacket(testPeerNode, 4242, gomeshproto.Routing_NONE)); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("unrelated routing packet produced output: %s", buf.String())
	}

	// Silence past the timeout is reported as a failure.
	buf.Reset()
	_ = s.execute(`{"id":"dm-3","to":"DG01","text":"hello?"}`)
	buf.Reset()
	base := s.now()
	s.now = func() time.Time { return base.Add(ackTimeout + time.Second) }
	if err := s.expirePending(); err != nil {
		t.Fatal(err)
	}
	expired := lines(buf)[0]
	if expired["type"] != "failed" || expired["id"] != "dm-3" || !strings.Contains(expired["error"].(string), "no acknowledgement") {
		t.Errorf("expired event = %v", expired)
	}
	if len(s.pending) != 0 {
		t.Errorf("pending should be empty after expiry, got %v", s.pending)
	}
}

func TestStreamNodeWithoutUserStillHasID(t *testing.T) {
	s, buf := newTestStream(t, false)
	// The radio lists some nodes with no user record at all.
	bare := uint32(463898580)
	if err := s.handle(&gomeshproto.FromRadio{PayloadVariant: &gomeshproto.FromRadio_NodeInfo{NodeInfo: &gomeshproto.NodeInfo{Num: bare}}}); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := s.handle(textPacket(bare, broadcastNodeNumber, 0, "😂")); err != nil {
		t.Fatal(err)
	}
	m := lines(buf)[0]
	if m["from"].(map[string]interface{})["node_id"] != "!1ba66ad4" {
		t.Errorf("from = %v", m["from"])
	}
	direct := m["reply"].(map[string]interface{})["direct"].(map[string]interface{})
	if direct["to"] != "!1ba66ad4" {
		t.Errorf("reply.direct should address the sender, got %v", direct)
	}
}

func TestStreamSendTextBuildsPacket(t *testing.T) {
	s, _ := newTestStream(t, false)
	if _, err := s.sendText(strings.Repeat("x", 238), 0, 0); err == nil {
		t.Error("oversized message should be rejected before reaching the radio")
	}
	id1, id2 := s.newPacketID(), s.newPacketID()
	if id1 == 0 || id2 == 0 || id1 == id2 {
		t.Errorf("packet ids should be non-zero and distinct, got %d and %d", id1, id2)
	}
}

func TestStreamCommands(t *testing.T) {
	s, buf := newTestStream(t, false)
	sender := &fakeSender{}
	s.send = sender.send

	cases := []struct {
		name     string
		line     string
		wantCall string // "" means nothing should be sent
		wantType string
		wantErr  string
	}{
		{"broadcast by default", `{"id":"a","channel":1,"text":"hello all"}`, "hello all|0|1", "sent", ""},
		{"numeric to", `{"to":1135973780,"text":"hi"}`, "hi|1135973780|0", "sent", ""},
		{"node id to", `{"to":"!43b59594","text":"hi"}`, "hi|1135973780|0", "sent", ""},
		{"long name to", `{"to":"dvdgmz01","text":"hi"}`, "hi|1135973780|0", "sent", ""},
		{"short name to", `{"to":"DG01","text":"hi"}`, "hi|1135973780|0", "sent", ""},
		{"reply template with text", `{"to":"!43b59594","channel":0,"text":"72F and sunny"}`, "72F and sunny|1135973780|0", "sent", ""},
		{"unknown name", `{"to":"nobody","text":"hi"}`, "", "error", `no known node named "nobody"`},
		{"bad id", `{"to":"!zz","text":"hi"}`, "", "error", `invalid node id "!zz"`},
		{"missing text", `{"to":"DG01"}`, "", "error", "text is required"},
		{"bad channel", `{"channel":9,"text":"hi"}`, "", "error", "channel must be between 0 and 7"},
		{"invalid json", `{"to":`, "", "error", "invalid command"},
		{"blank line", `   `, "", "", ""},
	}
	for _, c := range cases {
		buf.Reset()
		sender.calls = nil
		if err := s.execute(c.line); err != nil {
			t.Fatalf("%s: execute returned %v", c.name, err)
		}
		if c.wantType == "" {
			if buf.Len() != 0 {
				t.Errorf("%s: expected no output, got %s", c.name, buf.String())
			}
			continue
		}
		got := lines(buf)[0]
		if got["type"] != c.wantType {
			t.Errorf("%s: type = %v, want %v (%v)", c.name, got["type"], c.wantType, got["error"])
		}
		if c.wantErr != "" && !strings.Contains(got["error"].(string), c.wantErr) {
			t.Errorf("%s: error = %v, want it to contain %q", c.name, got["error"], c.wantErr)
		}
		if c.wantCall == "" && len(sender.calls) != 0 {
			t.Errorf("%s: unexpected send %v", c.name, sender.calls)
		}
		if c.wantCall != "" && (len(sender.calls) != 1 || sender.calls[0] != c.wantCall) {
			t.Errorf("%s: sends = %v, want [%s]", c.name, sender.calls, c.wantCall)
		}
	}

	// A sent event names the resolved recipient and echoes the id.
	buf.Reset()
	_ = s.execute(`{"id":"reply-17","to":"DG01","text":"hi"}`)
	got := lines(buf)[0]
	if got["id"] != "reply-17" || got["broadcast"] != false || got["to"].(map[string]interface{})["longname"] != "DvdGmz01" {
		t.Errorf("sent event = %v", got)
	}

	// Radio errors are relayed as error events, not failures of the stream.
	sender.err = errors.New("message too large")
	buf.Reset()
	if err := s.execute(`{"text":"x"}`); err != nil {
		t.Fatal(err)
	}
	if got := lines(buf)[0]; got["type"] != "error" || got["error"] != "message too large" {
		t.Errorf("relayed radio error = %v", got)
	}
}

func TestStreamStopsWhenStdinCloses(t *testing.T) {
	s, _ := newTestStream(t, false)
	lines := make(chan string)
	close(lines)
	s.commands = lines
	open, err := s.drainCommands()
	if err != nil || open {
		t.Errorf("drainCommands on closed stdin = open %v, err %v; want closed", open, err)
	}
}

func TestStreamOutgoingMessage(t *testing.T) {
	s, buf := newTestStream(t, false)
	buf.Reset()
	if err := s.handle(textPacket(testMyNode, testPeerNode, 0, "sent from the phone")); err != nil {
		t.Fatal(err)
	}
	m := lines(buf)[0]
	if m["outgoing"] != true || m["direct"] != false {
		t.Errorf("outgoing/direct = %v/%v", m["outgoing"], m["direct"])
	}
}
