package gmtcli

import (
	"testing"

	"github.com/lmatte7/meshtastic-go/pkg/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

func TestNodeDBResetPacket(t *testing.T) {
	out, err := nodeDBResetPacket(1397525059)
	if err != nil {
		t.Fatal(err)
	}

	var toRadio gomeshproto.ToRadio
	if err := proto.Unmarshal(out, &toRadio); err != nil {
		t.Fatal(err)
	}
	packet := toRadio.GetPacket()
	if packet.GetTo() != 1397525059 || !packet.GetWantAck() {
		t.Errorf("packet should be addressed to the radio itself with an ack, got %v", packet)
	}
	decoded := packet.GetDecoded()
	if decoded.GetPortnum() != gomeshproto.PortNum_ADMIN_APP || !decoded.GetWantResponse() {
		t.Errorf("payload should be an admin request wanting a response, got %v", decoded)
	}

	var admin gomeshproto.AdminMessage
	if err := proto.Unmarshal(decoded.GetPayload(), &admin); err != nil {
		t.Fatal(err)
	}
	if !admin.GetNodedbReset() {
		t.Errorf("admin message should be nodedb_reset, got %v", &admin)
	}
}
