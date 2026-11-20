package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"
)

func newTestBot() (*bot, *bytes.Buffer, *[]string) {
	var commands bytes.Buffer
	var lookups []string
	b := newBot(&commands, log.New(io.Discard, "", 0))
	clock := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	b.now = func() time.Time { return clock }
	b.weather = func(location string) (string, error) {
		lookups = append(lookups, location)
		return "☀️ Sunny +22°C (feels +21°C), wind ↗8km/h, humidity 20%", nil
	}
	return b, &commands, &lookups
}

func decodeLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("bot wrote invalid JSON %q: %v", line, err)
		}
		out = append(out, m)
	}
	return out
}

const weatherRequest = `{"type":"message","text":"!weather","direct":false,"outgoing":false,` +
	`"from":{"node_id":"!43b59594","longname":"DvdGmz01"},` +
	`"reply":{"direct":{"to":"!43b59594","channel":0},"channel":{"channel":0}}}`

// nodesEvent answers request id with a node list: our own node without GPS,
// two direct neighbours with positions, an MQTT node and a far node that must
// both be ignored, one node without a position, and one whose hop count the
// radio did not report, which must be ignored too.
func nodesEvent(id string, gps bool) string {
	me := `{"node_id":"!534c8a43","latitude_degrees":0,"longitude_degrees":0}`
	if gps {
		me = `{"node_id":"!534c8a43","latitude_degrees":36.1,"longitude_degrees":-115.2}`
	}
	return `{"type":"nodes","id":"` + id + `","my_node":` + me + `,"nodes":[` +
		`{"node_id":"!a","latitude_degrees":36.0,"longitude_degrees":-115.0,"hops_away":0,"via_mqtt":false,"last_heard":"t","last_heard_seconds_ago":600},` +
		`{"node_id":"!b","latitude_degrees":36.2,"longitude_degrees":-115.4,"hops_away":1,"via_mqtt":false,"last_heard":"t","last_heard_seconds_ago":1200},` +
		`{"node_id":"!c","latitude_degrees":40.0,"longitude_degrees":-100.0,"hops_away":0,"via_mqtt":true,"last_heard":"t","last_heard_seconds_ago":60},` +
		`{"node_id":"!d","latitude_degrees":10.0,"longitude_degrees":10.0,"hops_away":5,"via_mqtt":false,"last_heard":"t","last_heard_seconds_ago":60},` +
		`{"node_id":"!e","latitude_degrees":0,"longitude_degrees":0,"hops_away":0,"via_mqtt":false,"last_heard":"t","last_heard_seconds_ago":60},` +
		`{"node_id":"!f","latitude_degrees":50.0,"longitude_degrees":-120.0,"via_mqtt":false,"last_heard":"t","last_heard_seconds_ago":60}` +
		`]}`
}

func TestMessageLinesWithNumericIDsAreParsed(t *testing.T) {
	// Message lines carry the mesh packet id as a number; command results
	// carry our string ids. Both must parse.
	b, commands, _ := newTestBot()
	numeric := `{"type":"message","id":1836540931,"text":"!weather","direct":false,"outgoing":false,` +
		`"from":{"node_id":"!43b59594"},"reply":{"direct":{"to":"!43b59594","channel":0},"channel":{"channel":0}}}`
	if err := b.handle(numeric); err != nil {
		t.Fatalf("message line with numeric id: %v", err)
	}
	if got := decodeLines(t, commands); len(got) != 1 || got[0]["command"] != "nodes" {
		t.Errorf("request should have been accepted, got %v", got)
	}
	if err := b.handle(`{"type":"sent","id":"weather-1","text":"x"}`); err != nil {
		t.Fatalf("result line with string id: %v", err)
	}
}

func TestWeatherEstimatesPositionFromNearbyNodes(t *testing.T) {
	b, commands, lookups := newTestBot()
	if err := b.handle(weatherRequest); err != nil {
		t.Fatal(err)
	}
	got := decodeLines(t, commands)
	if len(got) != 1 || got[0]["command"] != "nodes" {
		t.Fatalf("expected a nodes request first, got %v", got)
	}
	id := got[0]["id"].(string)

	commands.Reset()
	if err := b.handle(nodesEvent(id, false)); err != nil {
		t.Fatal(err)
	}
	if len(*lookups) != 1 || (*lookups)[0] != "36.100,-115.200" {
		t.Errorf("lookup location = %v, want the mean of the two nearby nodes", *lookups)
	}
	reply := decodeLines(t, commands)[0]
	if reply["channel"] != float64(0) || reply["id"] != id {
		t.Errorf("reply = %v", reply)
	}
	if _, hasTo := reply["to"]; hasTo {
		t.Errorf("channel request should get a channel reply, got %v", reply)
	}
	// The reply is the report alone: no coordinates, node counts or notes.
	if text := reply["text"].(string); text != "☀️ Sunny +22°C (feels +21°C), wind ↗8km/h, humidity 20%" {
		t.Errorf("reply text = %q", text)
	}
}

func TestWeatherFallsBackToUnheardNodes(t *testing.T) {
	// Three nearby nodes, only one of them heard by the stream so far: too
	// few recent ones, so all three are used and the reply says so.
	b, commands, lookups := newTestBot()
	_ = b.handle(weatherRequest)
	id := decodeLines(t, commands)[0]["id"].(string)
	commands.Reset()
	stale := `{"type":"nodes","id":"` + id + `","my_node":{"node_id":"!534c8a43"},"nodes":[` +
		`{"node_id":"!a","latitude_degrees":36.0,"longitude_degrees":-115.0,"hops_away":0,"last_heard":"t","last_heard_seconds_ago":600},` +
		`{"node_id":"!b","latitude_degrees":36.2,"longitude_degrees":-115.2,"hops_away":1},` +
		`{"node_id":"!c","latitude_degrees":36.4,"longitude_degrees":-115.4,"hops_away":2}` +
		`]}`
	if err := b.handle(stale); err != nil {
		t.Fatal(err)
	}
	if (*lookups)[0] != "36.200,-115.200" {
		t.Errorf("lookup = %v, want the mean of all three nodes", *lookups)
	}
	if text := decodeLines(t, commands)[0]["text"].(string); strings.Contains(text, "nodes") {
		t.Errorf("reply must not mention how the position was found, got %q", text)
	}
}

func TestWeatherPrefersRadioGPS(t *testing.T) {
	b, commands, lookups := newTestBot()
	_ = b.handle(weatherRequest)
	id := decodeLines(t, commands)[0]["id"].(string)
	commands.Reset()
	if err := b.handle(nodesEvent(id, true)); err != nil {
		t.Fatal(err)
	}
	if (*lookups)[0] != "36.100,-115.200" || strings.Contains(decodeLines(t, commands)[0]["text"].(string), "GPS") {
		t.Errorf("lookups = %v, reply = %s (the reply must not mention the position source)", *lookups, commands.String())
	}
}

func TestWeatherNamedPlaceAndDirectReply(t *testing.T) {
	b, commands, lookups := newTestBot()
	direct := strings.Replace(weatherRequest, `"text":"!weather","direct":false`, `"text":"!Weather Reykjavik, Iceland","direct":true`, 1)
	_ = b.handle(direct)
	id := decodeLines(t, commands)[0]["id"].(string)
	commands.Reset()
	_ = b.handle(nodesEvent(id, false))
	if (*lookups)[0] != "Reykjavik, Iceland" {
		t.Errorf("lookup = %v", *lookups)
	}
	reply := decodeLines(t, commands)[0]
	if reply["to"] != "!43b59594" || strings.Contains(reply["text"].(string), "Reykjavik") {
		t.Errorf("direct reply = %v (the reply must not echo the place)", reply)
	}
}

func TestWeatherWithoutAnyPosition(t *testing.T) {
	b, commands, _ := newTestBot()
	_ = b.handle(weatherRequest)
	id := decodeLines(t, commands)[0]["id"].(string)
	commands.Reset()
	_ = b.handle(`{"type":"nodes","id":"` + id + `","my_node":{"node_id":"!534c8a43"},"nodes":[]}`)
	if text := decodeLines(t, commands)[0]["text"].(string); !strings.Contains(text, "don't know where I am") {
		t.Errorf("reply = %q", text)
	}
}

func TestWeatherLookupFailureAndThrottle(t *testing.T) {
	b, commands, lookups := newTestBot()
	b.weather = func(location string) (string, error) {
		*lookups = append(*lookups, location)
		return "", errors.New("boom")
	}
	_ = b.handle(weatherRequest)
	id := decodeLines(t, commands)[0]["id"].(string)
	commands.Reset()
	_ = b.handle(nodesEvent(id, false))
	if text := decodeLines(t, commands)[0]["text"].(string); !strings.Contains(text, "lookup failed") {
		t.Errorf("reply = %q", text)
	}

	// A flood of requests from other nodes within the interval: the first
	// sender is told the cooldown by direct message, the rest get nothing,
	// and no lookup happens.
	commands.Reset()
	for i := range 20 {
		other := strings.ReplaceAll(weatherRequest, "!43b59594", fmt.Sprintf("!%08x", 1000+i))
		_ = b.handle(other)
	}
	got := decodeLines(t, commands)
	if len(got) != 1 || len(*lookups) != 1 {
		t.Fatalf("expected exactly one cooldown notice and no lookups, got commands %q and %d lookups", commands.String(), len(*lookups))
	}
	if notice := got[0]; notice["to"] != "!000003e8" || notice["text"] != "!weather is available once every 5m0s." {
		t.Errorf("cooldown notice should go to the first turned-away sender by DM, got %v", notice)
	}

	// Once the interval has passed, the next request is served, and the
	// first throttled sender in the new period gets a notice again.
	base := b.now()
	b.now = func() time.Time { return base.Add(b.interval) }
	commands.Reset()
	_ = b.handle(weatherRequest)
	if got := decodeLines(t, commands); len(got) != 1 || got[0]["command"] != "nodes" {
		t.Errorf("request after the interval should be served, got %v", got)
	}
	commands.Reset()
	_ = b.handle(strings.ReplaceAll(weatherRequest, "!43b59594", "!00000777"))
	if got := decodeLines(t, commands); len(got) != 1 || got[0]["to"] != "!00000777" {
		t.Errorf("first throttled sender of the new period should get a notice, got %v", got)
	}
	commands.Reset()

	// Other lines never produce commands.
	for _, line := range []string{`{"type":"ready"}`, `{"type":"message","text":"!weather","outgoing":true,"reply":{"direct":{},"channel":{}}}`, `{"type":"sent","id":"weather-1"}`, `{"type":"nodes","id":"unknown","nodes":[]}`} {
		if err := b.handle(line); err != nil {
			t.Errorf("handle(%s): %v", line, err)
		}
	}
	if commands.Len() != 0 {
		t.Errorf("unexpected commands: %s", commands.String())
	}
}
