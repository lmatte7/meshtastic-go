package gomesh

import (
	"bytes"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	pb "github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
)

var port = flag.String("port", "", "port the radio is connected to")

func TestRadioInfo(t *testing.T) {

	radio := radioSetup(t)
	defer radio.Close()

	radioResponses, err := radio.GetRadioInfo()
	if err != nil {
		t.Fatalf("Error when communicating with radio: %v", err)
	}

	if len(radioResponses) < 4 {
		t.Fatalf("Missing Results")
	}

}

func TestGetChannelInfo(t *testing.T) {
	radio := radioSetup(t)
	defer radio.Close()

	channel, err := radio.GetChannelInfo(0)
	if err != nil {
		t.Fatalf("Error when retrieving with channel info: %v", err)
	}

	if channel.Index != 0 {
		t.Fatalf("Error when retrieving channel: %v", err)
	}
}

func TestSendText(t *testing.T) {

	radio := radioSetup(t)
	defer radio.Close()

	err := radio.SendTextMessage("Test", 0, 1)

	if err != nil {
		t.Fatalf("Error when communicating with radio: %v", err)
	}

}

func TestSetOwner(t *testing.T) {

	radio := radioSetup(t)
	defer radio.Close()

	err := radio.SetRadioOwner("Test Owner")
	if err != nil {
		t.Fatalf("Error when communicating with radio: %v", err)
	}

	radioResponses, err := radio.GetRadioInfo()
	if err != nil {
		t.Fatalf("Error when retrieving radio information: %v", err)
	}
	nodes := make([]*pb.FromRadio_NodeInfo, 0)
	var nodeNum uint32
	nodeNum = 0

	for _, response := range radioResponses {

		if nodeInfo, ok := response.GetPayloadVariant().(*pb.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}

		if info, ok := response.GetPayloadVariant().(*pb.FromRadio_MyInfo); ok {
			nodeNum = info.MyInfo.MyNodeNum
		}

	}

	for _, node := range nodes {
		if node.NodeInfo.Num == nodeNum {
			if node.NodeInfo.User.LongName != "Test Owner" {
				t.Fatalf("Owner name not correctly set")
			}
		}
	}
	// If test succeeds change name for future tests
	err = radio.SetRadioOwner("Owner")
	if err != nil {
		t.Fatalf("Error when communicating with radio: %v", err)
	}

}

func TestSetChannelURL(t *testing.T) {

	radio := radioSetup(t)
	defer radio.Close()

	err := radio.SetChannelURL("https://www.meshtastic.org/c/#CgoSBGFzZGY6AggNCigSIFUuaekdQ2K8pKQihn9HFxxTgY_QPvepwDvv7MDpFQ0EGgR0ZXN0")
	if err != nil {
		t.Fatalf("Error when communicating with radio: %v", err)
	}

	chanSettings, err := radio.GetChannelInfo(0)
	if err != nil {
		t.Fatalf("Error retreiving channel settings: %v", err)
	}

	psk := []byte("asdf")
	if !bytes.Equal(chanSettings.Settings.Psk, psk) {
		t.Fatalf("Channel PSK not set correctly from URL")
	}

	err = radio.DeleteChannel(1)
	if err != nil {
		t.Fatalf("Error Deleting channel: %v", err)
	}
}

func TestSetModem(t *testing.T) {

	radio := radioSetup(t)
	defer radio.Close()

	err := radio.SetModemMode("vls")
	if err != nil {
		t.Fatalf("Error when communicating with radio: %v", err)
	}

}

func TestSetRadioConfig(t *testing.T) {
	radio := radioSetup(t)
	defer radio.Close()

	err := radio.SetRadioConfig("DoubleTapAsButtonPress", "True")
	if err != nil {
		t.Fatalf("Error setting config: %v", err)
	}

	time.Sleep(3 * time.Second)
	configPackets, _, err := radio.GetRadioConfig()

	if err != nil {
		t.Fatalf("Error when opening serial communications with radio: %v", err)
	}

	for _, config := range configPackets {

		if device := config.Config.GetDevice(); device != nil {
			fmt.Printf("Config Packets: %v\n", device)
			if !device.DoubleTapAsButtonPress {
				t.Fatalf("Error setting config settings")
			}
		}
	}

}

func TestAddChannel(t *testing.T) {
	radio := radioSetup(t)
	defer radio.Close()

	// Clear out any existing channels used in testing
	_ = radio.DeleteChannel(1)
	time.Sleep(2 * time.Second)
	_ = radio.DeleteChannel(2)
	time.Sleep(2 * time.Second)
	_ = radio.DeleteChannel(3)
	time.Sleep(2 * time.Second)

	chanName := "test"
	chanNum := 2

	err := radio.AddChannel(chanName, chanNum)
	if err != nil {
		t.Fatalf("Error adding channel: %v", err)
	}

	time.Sleep(2 * time.Second)
	channel, err := radio.GetChannelInfo(chanNum)
	if err != nil {
		t.Fatalf("Error retrieving channel: %v", err)
	}

	if channel.Settings.Name != chanName {
		t.Fatalf("Failed to add channel")
	}

}

func TestDeleteChannel(t *testing.T) {
	radio := radioSetup(t)
	defer radio.Close()

	// Clear out any existing channels used in testing
	_ = radio.DeleteChannel(1)
	time.Sleep(2 * time.Second)
	_ = radio.DeleteChannel(2)
	time.Sleep(2 * time.Second)
	_ = radio.DeleteChannel(3)
	time.Sleep(2 * time.Second)

	chanName := "test"
	chanNum := 2

	err := radio.AddChannel(chanName, chanNum)
	if err != nil {
		t.Fatalf("Error adding channel: %v", err)
	}

	err = radio.DeleteChannel(chanNum)
	if err != nil {
		t.Fatalf("Error Deleting channel: %v", err)
	}

	time.Sleep(2 * time.Second)
	channel, err := radio.GetChannelInfo(chanNum)
	if err != nil {
		t.Fatalf("Error retrieving channel: %v", err)
	}

	if channel.Role != pb.Channel_DISABLED {
		t.Fatalf("Failed to delete channel")
	}
}

func TestSetChannelSettings(t *testing.T) {

	radio := radioSetup(t)
	defer radio.Close()

	// Clear out any existing channels used in testing
	_ = radio.DeleteChannel(1)
	time.Sleep(2 * time.Second)
	_ = radio.DeleteChannel(2)
	time.Sleep(2 * time.Second)
	_ = radio.DeleteChannel(3)
	time.Sleep(2 * time.Second)

	chanName := "test"
	chanNum := 2

	err := radio.AddChannel(chanName, chanNum)
	if err != nil {
		t.Fatalf("Error adding channel: %v", err)
	}

	time.Sleep(2 * time.Second)
	channel, err := radio.GetChannelInfo(chanNum)
	if err != nil {
		t.Fatalf("Error retrieving channel: %v", err)
	}

	if channel.Settings.Name != chanName {
		t.Fatalf("Failed to add channel")
	}

	newPsk := "newPsk"
	err = radio.SetChannel(chanNum, "Psk", newPsk)
	if err != nil {
		t.Fatalf("Error changing channel setting: %v", err)
	}

	time.Sleep(2 * time.Second)
	channel, err = radio.GetChannelInfo(chanNum)
	if err != nil {
		t.Fatalf("Error retrieving channel: %v", err)
	}

	if string(channel.Settings.Psk) != newPsk {
		t.Fatalf("Failed to change channel")
	}

}

func TestSetLocation(t *testing.T) {

	radio := radioSetup(t)
	defer radio.Close()

	latitude := int32(320602112)
	longitude := int32(-87818240)
	err := radio.SetLocation(latitude, longitude, 127)
	if err != nil {
		t.Fatalf("Error setting location: %v", err)
	}

	radioResponses, err := radio.GetRadioInfo()
	if err != nil {
		t.Fatalf("Error retreiving radio information: %v", err)
	}

	position := &gomeshproto.Position{}

	for _, packet := range radioResponses {
		if nodeInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			if nodeInfo.NodeInfo.Num == radio.nodeNum {
				position = nodeInfo.NodeInfo.Position
			}
		}
	}

	if position.GetLatitudeI() != latitude || position.GetLongitudeI() != longitude {
		t.Fatalf("Failed to set location")
	}
}

// radioSetup opens the radio named by the -port flag. The tests in this file
// need real hardware, so they are skipped when no port is given.
func radioSetup(t *testing.T) Radio {
	t.Helper()
	if *port == "" {
		t.Skip("no radio attached; pass -args -port=<serial port> to run hardware tests")
	}

	radio := Radio{}
	if err := radio.Init(*port); err != nil {
		t.Fatalf("Error when opening serial communications with radio: %v", err)
	}

	return radio
}
