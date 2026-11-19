package gmtcli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestLineEncoderWritesCompactLines(t *testing.T) {
	var buf bytes.Buffer
	enc := newLineEncoder(&buf)
	_ = enc.Encode(map[string]string{"type": "ready", "text": "<hi>"})
	_ = enc.Encode(map[string]int{"n": 1})
	got := buf.String()
	if strings.Count(got, "\n") != 2 || strings.Contains(got, "\n ") {
		t.Errorf("expected two compact lines, got %q", got)
	}
	if !strings.Contains(got, `"<hi>"`) {
		t.Errorf("HTML escaping should be off, got %q", got)
	}
}

func TestTerminalEncoderIndentsAndKeepsFieldOrder(t *testing.T) {
	// Colours would obscure the text comparison; NO_COLOR turns them off.
	t.Setenv("NO_COLOR", "1")
	defer func() { color.NoColor = false }()

	var buf bytes.Buffer
	enc := newTerminalEncoder(&buf)
	_ = enc.Encode(StreamReady{Type: "ready", Channels: []StreamChannel{{Index: 0, Name: "Default"}}})
	got := buf.String()
	if !strings.HasPrefix(got, "{\n    \"type\":\"ready\",\n") {
		t.Errorf("expected four-space indentation with type first, got %q", got)
	}
	if !strings.Contains(got, "\n        {\n            \"index\":0,") {
		t.Errorf("nested values should indent by four more spaces, got %q", got)
	}
	if strings.Contains(got, "\x1b[") {
		t.Errorf("NO_COLOR should disable colour, got %q", got)
	}
}

func TestStreamEncoderUsesLinesForPipes(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if isTerminal(w) {
		t.Fatal("a pipe should not count as a terminal")
	}
	if _, isLine := newStreamEncoder(w).(interface{ SetIndent(string, string) }); !isLine {
		t.Fatal("pipe output should use the plain line encoder")
	}
}
