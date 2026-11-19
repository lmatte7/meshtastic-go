package gmtcli

import (
	"testing"
	"time"

	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
)

func TestProtoNodeToMeshNode(t *testing.T) {
	// A fixed clock keeps the elapsed-time fields deterministic.
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	lastHeard := now.Add(-90 * time.Second).Unix()

	node := &gomeshproto.FromRadio_NodeInfo{
		NodeInfo: &gomeshproto.NodeInfo{
			Num:       2617543978,
			LastHeard: uint32(lastHeard),
			User: &gomeshproto.User{
				Id:        "!9c048d2a",
				LongName:  "PalaceStation",
				ShortName: "PlcS",
				Macaddr:   []byte{0xd5, 0x0c, 0x9c, 0x04, 0x8d, 0x2a},
				HwModel:   gomeshproto.HardwareModel_RAK4631,
				Role:      gomeshproto.Config_DeviceConfig_CLIENT_MUTE,
			},
			DeviceMetrics: &gomeshproto.DeviceMetrics{
				BatteryLevel:       100,
				Voltage:            4.2,
				ChannelUtilization: 18.5,
				AirUtilTx:          2.9,
			},
			Position: &gomeshproto.Position{
				LatitudeI:  361758720,
				LongitudeI: -1152385024,
				Altitude:   750,
			},
		},
	}

	mn := protoNodeToMeshNodeAt(node, now)

	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"ID", mn.ID, "!9c048d2a"},
		{"NodeNumber", mn.NodeNumber, uint32(2617543978)},
		{"LongName", mn.LongName, "PalaceStation"},
		{"ShortName", mn.ShortName, "PlcS"},
		{"MACAddress", mn.MACAddress, "d5:0c:9c:04:8d:2a"},
		{"HardwareModel", mn.HardwareModel, "RAK4631"},
		{"Role", mn.Role, "CLIENT_MUTE"},
		{"BatteryLevel", mn.BatteryLevel, 100},
		{"AltitudeMeters", mn.AltitudeMeters, 750},
		{"LatitudeDegrees", mn.LatitudeDegrees, 36.175872},
		{"LongitudeDegrees", mn.LongitudeDegrees, -115.2385024},
		{"LastHeard", mn.LastHeard, time.Unix(lastHeard, 0).UTC().Format(time.RFC3339)},
		{"LastHeardHuman", mn.LastHeardHuman, "1m30s"},
		{"LastHeardSecondsAgo", mn.LastHeardSecondsAgo, 90},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestProtoNodeToMeshNodeSparse(t *testing.T) {
	// A node with no user, metrics or position must not panic and must leave
	// the optional fields empty so omitempty drops them from the JSON.
	mn := ProtoNodeToMeshNode(&gomeshproto.FromRadio_NodeInfo{
		NodeInfo: &gomeshproto.NodeInfo{Num: 42},
	})
	if mn.NodeNumber != 42 {
		t.Errorf("NodeNumber = %d, want 42", mn.NodeNumber)
	}
	if mn.ID != "" || mn.Role != "" || mn.HardwareModel != "" || mn.MACAddress != "" {
		t.Errorf("expected empty user fields, got %+v", mn)
	}
	if mn.LastHeard != "" || mn.LastHeardHuman != "" || mn.LastHeardSecondsAgo != 0 {
		t.Errorf("expected empty last-heard fields, got %+v", mn)
	}

	if empty := ProtoNodeToMeshNode(&gomeshproto.FromRadio_NodeInfo{}); empty != (MeshNode{}) {
		t.Errorf("expected zero MeshNode for empty packet, got %+v", empty)
	}
}
