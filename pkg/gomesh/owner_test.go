package gomesh

import (
	"testing"
	"unicode/utf8"
)

func TestDefaultShortName(t *testing.T) {
	cases := map[string]string{
		"Alice's T-Echo": "Alic",
		"ab":             "ab",
		"東京":             "東",  // 3 bytes; the second character would make 6
		"éé":             "éé", // 2 bytes each, exactly the limit
		"éééé":           "éé",
		"Ünïcode node":   "Ün", // Ü (2) + n (1) = 3; adding ï (2) would exceed 4
	}
	for name, want := range cases {
		got := defaultShortName(name)
		if got != want {
			t.Errorf("defaultShortName(%q) = %q, want %q", name, got, want)
		}
		if len(got) > maxShortNameBytes || !utf8.ValidString(got) {
			t.Errorf("defaultShortName(%q) = %q is over the limit or not valid UTF-8", name, got)
		}
	}
}

func TestSetRadioOwnerNamesRejectsLongShortName(t *testing.T) {
	radio, _ := testRadio()
	if err := radio.SetRadioOwnerNames("Some Node", "TOOLONG"); err == nil {
		t.Error("a short name over four bytes should be rejected")
	}
	if err := radio.SetRadioOwnerNames("Some Node", "東京"); err == nil {
		t.Error("a six-byte short name should be rejected even though it is two characters")
	}
}
