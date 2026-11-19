package gmtcli

import (
	"strings"
	"testing"

	"github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

func TestTracerouteRequest(t *testing.T) {
	out, err := tracerouteRequest(1135973780, 4242)
	if err != nil {
		t.Fatal(err)
	}
	var toRadio gomeshproto.ToRadio
	if err := proto.Unmarshal(out, &toRadio); err != nil {
		t.Fatal(err)
	}
	packet := toRadio.GetPacket()
	decoded := packet.GetDecoded()
	if packet.GetTo() != 1135973780 || packet.GetId() != 4242 || decoded.GetPortnum() != gomeshproto.PortNum_TRACEROUTE_APP || !decoded.GetWantResponse() {
		t.Errorf("request = %v", packet)
	}
}

func TestBuildAndFormatTraceroute(t *testing.T) {
	nodes := map[uint32]MeshNode{
		1: {ID: "!00000001", LongName: "Me", ShortName: "ME"},
		2: {ID: "!00000002", LongName: "Relay", ShortName: "RLY"},
		3: {ID: "!00000003", LongName: "Far Away", ShortName: "FAR"},
	}
	discovery := &gomeshproto.RouteDiscovery{
		Route:      []uint32{2},
		SnrTowards: []int32{21, -12}, // 5.25 dB into the relay, -3 dB into the destination
		RouteBack:  []uint32{2},
		SnrBack:    []int32{unknownSNR, 8},
	}
	tr := buildTraceroute(1, 3, discovery, nodes)

	if tr.From.ID != "!00000001" || tr.To.LongName != "Far Away" {
		t.Errorf("endpoints = %+v / %+v", tr.From, tr.To)
	}
	if len(tr.Towards) != 2 || tr.Towards[0].ID != "!00000002" || *tr.Towards[0].SNR != 5.25 || tr.Towards[1].ID != "!00000003" || *tr.Towards[1].SNR != -3 {
		t.Errorf("towards = %+v", tr.Towards)
	}
	if len(tr.Back) != 2 || tr.Back[0].SNR != nil || *tr.Back[1].SNR != 2 || tr.Back[1].ID != "!00000001" {
		t.Errorf("back = %+v", tr.Back)
	}

	text := formatTraceroute(tr)
	for _, want := range []string{"Route to Far Away (!00000003), 2 hop(s):", "Me (!00000001)", "+5.25 dB", "Relay (!00000002)", "-3.00 dB", "Route back, 2 hop(s):", "? dB", "+2.00 dB"} {
		if !strings.Contains(text, want) {
			t.Errorf("formatted route missing %q:\n%s", want, text)
		}
	}

	// A direct reply: no relays, one link each way, unknown node named by id.
	direct := buildTraceroute(1, 9, &gomeshproto.RouteDiscovery{SnrTowards: []int32{40}}, nodes)
	if len(direct.Towards) != 1 || direct.Towards[0].ID != "!00000009" || *direct.Towards[0].SNR != 10 || direct.Back != nil {
		t.Errorf("direct = %+v", direct)
	}

	// A relay the firmware could not identify is recorded as the broadcast
	// address and labelled as unknown.
	relayed := buildTraceroute(1, 3, &gomeshproto.RouteDiscovery{Route: []uint32{broadcastNodeNumber}, SnrTowards: []int32{unknownSNR, 8}}, nodes)
	if relayed.Towards[0].ID != "!ffffffff" || relayed.Towards[0].LongName != "unknown relay" || relayed.Towards[0].SNR != nil {
		t.Errorf("unknown relay hop = %+v", relayed.Towards[0])
	}
	if !strings.Contains(formatTraceroute(relayed), "unknown relay (!ffffffff)") {
		t.Errorf("formatted output should label the unknown relay:\n%s", formatTraceroute(relayed))
	}
}

func TestLookupNode(t *testing.T) {
	nodes := map[uint32]MeshNode{1135973780: {ID: "!43b59594", LongName: "DvdGmz01", ShortName: "DG01"}}
	for _, ref := range []string{"1135973780", "!43b59594", "DvdGmz01", "dg01", " DG01 "} {
		if got, err := lookupNode(ref, nodes); err != nil || got != 1135973780 {
			t.Errorf("lookupNode(%q) = %d, %v", ref, got, err)
		}
	}
	for _, ref := range []string{"!zz", "nobody"} {
		if _, err := lookupNode(ref, nodes); err == nil {
			t.Errorf("lookupNode(%q) should fail", ref)
		}
	}
}
