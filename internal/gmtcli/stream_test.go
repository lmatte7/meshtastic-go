package gmtcli

import (
	"bytes"
	"encoding/json"
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
