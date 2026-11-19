package gmtcli

import (
	"testing"
	"time"

	"github.com/lmatte7/meshtastic-go/internal/gomesh"
	"github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

func TestProtoNodeToMeshNode(t *testing.T) {
	// A fixed clock keeps the elapsed-time fields deterministic.
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	lastHeard := now.Add(-90 * time.Second).Unix()

	node := &gomeshproto.FromRadio_NodeInfo{
		NodeInfo: &gomeshproto.NodeInfo{
			Num:       2617543978,
			LastHeard: uint32(lastHeard),
			Snr:       -1.75,
			ViaMqtt:   true,
			User: &gomeshproto.User{
				Id:        "!9c048d2a",
				LongName:  "PalaceStation",
				ShortName: "PlcS",
				Macaddr:   []byte{0xd5, 0x0c, 0x9c, 0x04, 0x8d, 0x2a},
				HwModel:   gomeshproto.HardwareModel_RAK4631,
				Role:      gomeshproto.Config_DeviceConfig_CLIENT_MUTE,
			},
			DeviceMetrics: &gomeshproto.DeviceMetrics{
				BatteryLevel:       proto.Uint32(100),
				Voltage:            proto.Float32(4.2169998),
				ChannelUtilization: proto.Float32(18.366666793823242),
				AirUtilTx:          proto.Float32(2.9119999408721924),
				UptimeSeconds:      proto.Uint32(3600),
			},
			Position: &gomeshproto.Position{
				LatitudeI:  proto.Int32(361758720),
				LongitudeI: proto.Int32(-1152385024),
				Altitude:   proto.Int32(750),
			},
			HopsAway: proto.Uint32(2),
		},
	}

	mn := protoNodeToMeshNodeAt(node, now)

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"ID", mn.ID, "!9c048d2a"},
		{"NodeNumber", mn.NodeNumber, uint32(2617543978)},
		{"LongName", mn.LongName, "PalaceStation"},
		{"ShortName", mn.ShortName, "PlcS"},
		{"MACAddress", mn.MACAddress, "d5:0c:9c:04:8d:2a"},
		{"HardwareModel", mn.HardwareModel, "RAK4631"},
		{"Role", mn.Role, "CLIENT_MUTE"},
		{"BatteryLevel", mn.BatteryLevel, 100},
		{"Voltage", mn.Voltage, 4.217},
		{"ChannelUtilization", mn.ChannelUtilization, 18.37},
		{"AirUtilTx", mn.AirUtilTx, 2.91},
		{"AltitudeMeters", mn.AltitudeMeters, 750},
		{"LatitudeDegrees", mn.LatitudeDegrees, 36.175872},
		{"LongitudeDegrees", mn.LongitudeDegrees, -115.2385024},
		{"LastHeard", mn.LastHeard, time.Unix(lastHeard, 0).UTC().Format(time.RFC3339)},
		{"LastHeardHuman", mn.LastHeardHuman, "1m30s"},
		{"LastHeardSecondsAgo", mn.LastHeardSecondsAgo, 90},
		{"SNR", mn.SNR, -1.75},
		{"ViaMQTT", mn.ViaMQTT, true},
		{"UptimeSeconds", mn.UptimeSeconds, 3600},
	}
	if mn.HopsAway == nil || *mn.HopsAway != 2 {
		t.Errorf("HopsAway = %v, want 2", mn.HopsAway)
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestParseModemPreset(t *testing.T) {
	cases := map[string]gomeshproto.Config_LoRaConfig_ModemPreset{
		"lf":           gomeshproto.Config_LoRaConfig_LONG_FAST,
		"LF":           gomeshproto.Config_LoRaConfig_LONG_FAST,
		"st":           gomeshproto.Config_LoRaConfig_SHORT_TURBO,
		"LONG_FAST":    gomeshproto.Config_LoRaConfig_LONG_FAST,
		"medium_turbo": gomeshproto.Config_LoRaConfig_MEDIUM_TURBO,
	}
	for in, want := range cases {
		got, err := gomesh.ParseModemPreset(in)
		if err != nil || got != want {
			t.Errorf("ParseModemPreset(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	if _, err := gomesh.ParseModemPreset("bogus"); err == nil {
		t.Error("ParseModemPreset(\"bogus\") succeeded, want error")
	}
}

func TestHumanDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "0s"},
		{45 * time.Second, "45s"},
		{90 * time.Second, "1m30s"},
		{9*time.Hour + 5*time.Minute + 18*time.Second, "9h5m18s"},
		{104*time.Hour + 36*time.Minute + 34*time.Second, "4d8h36m"},
		{-5 * time.Second, "0s"},
	}
	for _, c := range cases {
		if got := humanDuration(c.d); got != c.want {
			t.Errorf("humanDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestProtoNodeToMeshNodeSparse(t *testing.T) {
	// A node with no user, metrics or position must not panic and must leave
	// the optional fields empty so omitempty drops them from the JSON.
	mn := ProtoNodeToMeshNode(&gomeshproto.FromRadio_NodeInfo{
		NodeInfo: &gomeshproto.NodeInfo{Num: 42},
	})
	if mn.NodeNumber != 42 || mn.ID != "!0000002a" {
		t.Errorf("NodeNumber/ID = %d/%q, want 42/!0000002a", mn.NodeNumber, mn.ID)
	}
	if mn.LongName != "" || mn.Role != "" || mn.HardwareModel != "" || mn.MACAddress != "" {
		t.Errorf("expected empty user fields, got %+v", mn)
	}
	if mn.LastHeard != "" || mn.LastHeardHuman != "" || mn.LastHeardSecondsAgo != 0 {
		t.Errorf("expected empty last-heard fields, got %+v", mn)
	}
	if mn.HopsAway != nil {
		t.Errorf("expected nil HopsAway when unreported, got %d", *mn.HopsAway)
	}

	if empty := ProtoNodeToMeshNode(&gomeshproto.FromRadio_NodeInfo{}); empty != (MeshNode{}) {
		t.Errorf("expected zero MeshNode for empty packet, got %+v", empty)
	}
}
