package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strings"
	"testing"
)

func newTestBot() (*bot, *bytes.Buffer) {
	var commands bytes.Buffer
	return &bot{commands: &commands, log: log.New(io.Discard, "", 0)}, &commands
}

const channelPing = `{"type":"message","id":1836540931,"text":"!ping","direct":false,"outgoing":false,` +
	`"from":{"node_id":"!43b59594","longname":"DvdGmz01"},` +
	`"reply":{"direct":{"to":"!43b59594","channel":1},"channel":{"channel":1}}}`

const directPing = `{"type":"message","text":" !PING ","direct":true,"outgoing":false,` +
	`"from":{"node_id":"!43b59594"},` +
	`"reply":{"direct":{"to":"!43b59594","channel":0},"channel":{"channel":0}}}`

func decode(t *testing.T, line string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("bot wrote invalid JSON %q: %v", line, err)
	}
	return m
}

func TestChannelPingIsAnsweredOnTheChannel(t *testing.T) {
	b, commands := newTestBot()
	if err := b.handle(channelPing); err != nil {
		t.Fatal(err)
	}
	cmd := decode(t, strings.TrimSpace(commands.String()))
	if cmd["text"] != "pong" || cmd["channel"] != float64(1) || cmd["id"] != "pong-1" {
		t.Errorf("command = %v", cmd)
	}
	if _, hasTo := cmd["to"]; hasTo {
		t.Errorf("channel reply must not address a single node, got %v", cmd)
	}
}

func TestDirectPingIsAnsweredDirectly(t *testing.T) {
	b, commands := newTestBot()
	if err := b.handle(directPing); err != nil {
		t.Fatal(err)
	}
	cmd := decode(t, strings.TrimSpace(commands.String()))
	if cmd["to"] != "!43b59594" || cmd["text"] != "pong" {
		t.Errorf("command = %v", cmd)
	}
}

func TestOtherLinesProduceNoCommands(t *testing.T) {
	b, commands := newTestBot()
	lines := []string{
		`{"type":"ready","channels":[]}`,
		`{"type":"message","text":"hello","direct":false,"reply":{"direct":{},"channel":{}}}`,
		`{"type":"message","text":"!ping","outgoing":true,"reply":{"direct":{},"channel":{}}}`,
		`{"type":"packet","port":"POSITION_APP"}`,
		`{"type":"sent","id":"pong-1"}`,
		`{"type":"delivered","id":"pong-1","implicit":true}`,
		`{"type":"failed","id":"pong-1","error":"MAX_RETRANSMIT"}`,
	}
	for _, line := range lines {
		if err := b.handle(line); err != nil {
			t.Errorf("handle(%s): %v", line, err)
		}
	}
	if commands.Len() != 0 {
		t.Errorf("expected no commands, got %s", commands.String())
	}
	if err := b.handle("not json"); err == nil {
		t.Error("unreadable lines should be reported")
	}
}
