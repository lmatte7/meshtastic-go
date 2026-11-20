package gmtcli

import (
	"testing"

	"github.com/lmatte7/meshtastic-go/pkg/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

func TestRebootPacket(t *testing.T) {
	out, err := rebootPacket(1397525059, 1)
	if err != nil {
		t.Fatal(err)
	}
	var toRadio gomeshproto.ToRadio
	if err := proto.Unmarshal(out, &toRadio); err != nil {
		t.Fatal(err)
	}
	packet := toRadio.GetPacket()
	if packet.GetTo() != 1397525059 || packet.GetDecoded().GetPortnum() != gomeshproto.PortNum_ADMIN_APP {
		t.Errorf("packet = %v", packet)
	}
	var admin gomeshproto.AdminMessage
	if err := proto.Unmarshal(packet.GetDecoded().GetPayload(), &admin); err != nil {
		t.Fatal(err)
	}
	if admin.GetRebootSeconds() != 1 {
		t.Errorf("admin message should request a reboot in 1s, got %v", &admin)
	}
}
