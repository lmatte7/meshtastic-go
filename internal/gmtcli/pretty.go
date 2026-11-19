package gmtcli

import (
	"encoding/json"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/nwidger/jsoncolor"
)

// eventEncoder writes one JSON document per Encode call.
type eventEncoder interface {
	Encode(v interface{}) error
}

// newStreamEncoder picks the stream's output format: strict JSON Lines when
// stdout is a pipe or file, so other programs can parse it, and indented,
// colourised JSON when it is a terminal, so a person can read it.
func newStreamEncoder(f *os.File) eventEncoder {
	if isTerminal(f) {
		return newTerminalEncoder(f)
	}
	return newLineEncoder(f)
}

// newLineEncoder writes one compact JSON document per line.
func newLineEncoder(w io.Writer) eventEncoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

// newTerminalEncoder writes indented JSON with colours. NO_COLOR in the
// environment turns the colours off.
func newTerminalEncoder(w io.Writer) eventEncoder {
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
	enc := jsoncolor.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)
	return enc
}

// isTerminal reports whether f is a character device such as a terminal.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
