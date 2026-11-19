package gomesh

import (
	"reflect"
	"strings"
	"testing"

	pb "github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
)

func TestSetFieldValue(t *testing.T) {
	lora := &pb.Config_LoRaConfig{}
	device := &pb.Config_DeviceConfig{}
	position := &pb.Config_PositionConfig{}

	cases := []struct {
		msg   any
		key   string
		value string
		check func() bool
	}{
		{lora, "Region", "US", func() bool { return lora.Region == pb.Config_LoRaConfig_US }},
		{lora, "Region", "eu_868", func() bool { return lora.Region == pb.Config_LoRaConfig_EU_868 }},
		{lora, "Region", "3", func() bool { return lora.Region == pb.Config_LoRaConfig_EU_868 }},
		{lora, "ModemPreset", "LONG_FAST", func() bool { return lora.ModemPreset == pb.Config_LoRaConfig_LONG_FAST }},
		{lora, "HopLimit", "3", func() bool { return lora.HopLimit == 3 }},
		{lora, "TxEnabled", "true", func() bool { return lora.TxEnabled }},
		{device, "Role", "CLIENT_MUTE", func() bool { return device.Role == pb.Config_DeviceConfig_CLIENT_MUTE }},
		{device, "Tzdef", "PST8PDT", func() bool { return device.Tzdef == "PST8PDT" }},
		{position, "GpsMode", "ENABLED", func() bool { return position.GpsMode == pb.Config_PositionConfig_ENABLED }},
	}
	for _, c := range cases {
		if err := setProtoValue(c.msg, c.key, c.value); err != nil {
			t.Errorf("set %s=%q: %v", c.key, c.value, err)
			continue
		}
		if !c.check() {
			t.Errorf("set %s=%q did not take effect", c.key, c.value)
		}
	}

	err := setProtoValue(lora, "Region", "MARS")
	if err == nil || !strings.Contains(err.Error(), "choose one of") || !strings.Contains(err.Error(), "EU_868") {
		t.Errorf("unknown enum name should list the choices, got %v", err)
	}
	if err := setProtoValue(lora, "HopLimit", "lots"); err == nil {
		t.Error("non-numeric value for a number should fail")
	}
	if err := setFieldValue(reflect.Value{}, "x"); err == nil {
		t.Error("an invalid field should fail")
	}
}
