package gomesh

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	pb "github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

const start1 = byte(0x94)
const start2 = byte(0xc3)
const headerLen = 4
const maxToFromRadioSize = 512
const broadcastAddr = "^all"
const localAddr = "^local"
const defaultHopLimit = 3
const broadcastNum = 0xffffffff

// configRequestID is sent with want_config_id; the radio echoes it back in
// config_complete_id once it has finished sending its configuration and node
// database.
const configRequestID = 42

// configTimeout bounds how long GetRadioInfo waits for the radio to finish
// answering a configuration request. ESP32 boards reboot when the serial port
// is opened and need several seconds before they respond at all.
const configTimeout = 30 * time.Second

// configRetryInterval is how often the configuration request is repeated
// while the radio has not answered yet.
const configRetryInterval = 3 * time.Second

// Radio holds the port and serial io.ReadWriteCloser struct to maintain one serial connection
type Radio struct {
	streamer streamer
	nodeNum  uint32
}

// Init initializes the Serial connection for the radio
func (r *Radio) Init(port string) error {

	streamer := streamer{}
	err := streamer.Init(port)
	if err != nil {
		return err
	}
	r.streamer = streamer

	err = r.getNodeNum()
	if err != nil {
		return err
	}

	return nil
}

// SendPacket takes a protbuf packet, construct the appropriate header and sends it to the radio
func (r *Radio) SendPacket(protobufPacket []byte) (err error) {

	packageLength := len(protobufPacket)

	header := []byte{start1, start2, byte(packageLength>>8) & 0xff, byte(packageLength) & 0xff}

	radioPacket := append(header, protobufPacket...)
	err = r.streamer.Write(radioPacket)
	if err != nil {
		return err
	}

	return

}

// ReadResponse reads whatever the radio is currently sending, decodes it into
// FromRadio packets and returns once the link goes idle.
//
// Each packet arrives framed as [START1, START2, length (2 bytes, big endian),
// protobuf]. Bytes outside a frame (for example boot or debug text) are
// skipped until the next START1/START2 pair. The timeout argument is kept for
// compatibility; the read always returns when the link is idle.
func (r *Radio) ReadResponse(timeout bool) (packets []*pb.FromRadio, err error) {
	b := make([]byte, 1)
	var frame []byte

	for {
		n, err := r.streamer.Read(b)
		if n == 0 || errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded) {
			// Idle: the serial inter-character timeout or TCP read deadline
			// expired without data.
			return packets, nil
		}
		if err != nil {
			return packets, err
		}

		frame = append(frame, b[0])

		switch {
		case len(frame) == 1 && frame[0] != start1:
			frame = frame[:0]
		case len(frame) == 2 && frame[1] != start2:
			frame = frame[:0]
		case len(frame) == headerLen && frameLength(frame) > maxToFromRadioSize:
			// Not a real header; resynchronise on the next start byte.
			frame = frame[:0]
		case len(frame) > headerLen && len(frame) == headerLen+frameLength(frame):
			packet := &pb.FromRadio{}
			if err := proto.Unmarshal(frame[headerLen:], packet); err == nil {
				packets = append(packets, packet)
			}
			frame = frame[:0]
		}
	}
}

// frameLength returns the payload length encoded in a frame header.
func frameLength(frame []byte) int {
	return int(frame[2])<<8 | int(frame[3])
}

// createAdminPacket builds a admin message packet to send to the radio
func (r *Radio) createAdminPacket(nodeNum uint32, payload []byte) (packetOut []byte, err error) {

	radioMessage := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: &pb.MeshPacket{
				To:      nodeNum,
				WantAck: true,
				PayloadVariant: &pb.MeshPacket_Decoded{
					Decoded: &pb.Data{
						Payload:      payload,
						Portnum:      pb.PortNum_ADMIN_APP,
						WantResponse: true,
					},
				},
			},
		},
	}

	packetOut, err = proto.Marshal(&radioMessage)
	if err != nil {
		return nil, err
	}

	return

}

// getNodeNum queries the radio and stores its own node number, which admin
// messages are addressed to.
func (r *Radio) getNodeNum() error {
	radioResponses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	for _, response := range radioResponses {
		if info, ok := response.GetPayloadVariant().(*pb.FromRadio_MyInfo); ok {
			r.nodeNum = info.MyInfo.MyNodeNum
		}
	}

	return nil
}

// GetRadioInfo asks the radio for its configuration and node database and
// returns every packet it sends up to the matching config_complete_id.
//
// The request is repeated every few seconds until the radio starts answering,
// because boards that reboot when the serial port is opened discard anything
// sent before their firmware is up. Reading continues across short gaps in
// the stream until the radio reports the dump complete or the overall timeout
// expires.
func (r *Radio) GetRadioInfo() ([]*pb.FromRadio, error) {
	request := pb.ToRadio{PayloadVariant: &pb.ToRadio_WantConfigId{WantConfigId: configRequestID}}
	out, err := proto.Marshal(&request)
	if err != nil {
		return nil, err
	}

	var responses []*pb.FromRadio
	deadline := time.Now().Add(configTimeout)
	var nextRequest time.Time

	for time.Now().Before(deadline) {
		if len(responses) == 0 && !time.Now().Before(nextRequest) {
			if err := r.SendPacket(out); err != nil {
				return nil, err
			}
			nextRequest = time.Now().Add(configRetryInterval)
		}

		packets, err := r.ReadResponse(true)
		if err != nil {
			return nil, err
		}

		for _, packet := range packets {
			responses = append(responses, packet)
			if packet.GetConfigCompleteId() == configRequestID {
				return responses, nil
			}
		}
	}

	if len(responses) == 0 {
		return nil, errors.New("no response from radio")
	}

	return responses, nil
}

// SendTextMessage sends a free form text message to other radios
func (r *Radio) SendTextMessage(message string, to int64, channel int64) error {
	var address int64
	if to == 0 {
		address = broadcastNum
	} else {
		address = to
	}

	// This constant is defined in Constants_DATA_PAYLOAD_LEN, but not in a friendly way to use
	if len(message) > 240 {
		return errors.New("message too large")
	}

	packetID := rand.Intn(2386828-1) + 1

	radioMessage := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: &pb.MeshPacket{
				To:      uint32(address),
				WantAck: true,
				Id:      uint32(packetID),
				Channel: uint32(channel),
				PayloadVariant: &pb.MeshPacket_Decoded{
					Decoded: &pb.Data{
						Payload: []byte(message),
						Portnum: pb.PortNum_TEXT_MESSAGE_APP,
					},
				},
			},
		},
	}

	out, err := proto.Marshal(&radioMessage)
	if err != nil {
		return err
	}

	if err := r.SendPacket(out); err != nil {
		return err
	}

	return nil

}

// SetRadioOwner sets the owner of the radio visible on the public mesh
func (r *Radio) SetRadioOwner(name string) error {

	if len(name) <= 2 {
		return errors.New("name too short")
	}

	adminPacket := pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetOwner{
			SetOwner: &pb.User{
				LongName:  name,
				ShortName: name[:3],
			},
		},
	}

	out, err := proto.Marshal(&adminPacket)
	if err != nil {
		return err
	}

	nodeNum := r.nodeNum

	packet, err := r.createAdminPacket(nodeNum, out)
	if err != nil {
		return err
	}

	if err := r.SendPacket(packet); err != nil {
		return err
	}

	return nil
}

// modemShortCodes are the short aliases accepted for modem presets alongside
// the full preset names.
var modemShortCodes = map[string]pb.Config_LoRaConfig_ModemPreset{
	"lf":  pb.Config_LoRaConfig_LONG_FAST,
	"ls":  pb.Config_LoRaConfig_LONG_SLOW,
	"vls": pb.Config_LoRaConfig_VERY_LONG_SLOW,
	"lm":  pb.Config_LoRaConfig_LONG_MODERATE,
	"lt":  pb.Config_LoRaConfig_LONG_TURBO,
	"ms":  pb.Config_LoRaConfig_MEDIUM_SLOW,
	"mf":  pb.Config_LoRaConfig_MEDIUM_FAST,
	"mt":  pb.Config_LoRaConfig_MEDIUM_TURBO,
	"sl":  pb.Config_LoRaConfig_SHORT_SLOW,
	"sf":  pb.Config_LoRaConfig_SHORT_FAST,
	"st":  pb.Config_LoRaConfig_SHORT_TURBO,
}

// ModemPresetShortCode returns the short alias for a preset, or "" if it has
// none.
func ModemPresetShortCode(preset pb.Config_LoRaConfig_ModemPreset) string {
	for code, p := range modemShortCodes {
		if p == preset {
			return code
		}
	}
	return ""
}

// ParseModemPreset resolves a short code such as "lf" or a preset name such
// as "LONG_FAST" (case-insensitive) to a modem preset.
func ParseModemPreset(mode string) (pb.Config_LoRaConfig_ModemPreset, error) {
	if preset, ok := modemShortCodes[strings.ToLower(mode)]; ok {
		return preset, nil
	}
	if value, ok := pb.Config_LoRaConfig_ModemPreset_value[strings.ToUpper(mode)]; ok {
		return pb.Config_LoRaConfig_ModemPreset(value), nil
	}
	return 0, fmt.Errorf("unknown modem preset %q", mode)
}

// SetModemMode sets the LoRa modem preset. The mode is a short code or a
// preset name as accepted by ParseModemPreset. The rest of the LoRa
// configuration (region, hop limit, ...) is read from the radio first and
// sent back unchanged.
func (r *Radio) SetModemMode(mode string) error {
	modemSetting, err := ParseModemPreset(mode)
	if err != nil {
		return err
	}

	configs, _, err := r.GetRadioConfig()
	if err != nil {
		return err
	}

	var lora *pb.Config_LoRaConfig
	for _, config := range configs {
		if l := config.Config.GetLora(); l != nil {
			lora = l
		}
	}
	if lora == nil {
		return errors.New("radio did not report its LoRa config")
	}

	lora.UsePreset = true
	lora.ModemPreset = modemSetting

	adminPacket := pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetConfig{
			SetConfig: &pb.Config{
				PayloadVariant: &pb.Config_Lora{
					Lora: lora,
				},
			},
		},
	}

	out, err := proto.Marshal(&adminPacket)
	if err != nil {
		return err
	}

	nodeNum := r.nodeNum

	packet, err := r.createAdminPacket(nodeNum, out)
	if err != nil {
		return err
	}

	if err := r.SendPacket(packet); err != nil {
		return err
	}

	return nil

}

// SetLocation sets a fixed location for the radio
func (r *Radio) SetLocation(lat int32, long int32, alt int32) error {

	positionPacket := pb.Position{
		LatitudeI:  proto.Int32(lat),
		LongitudeI: proto.Int32(long),
		Altitude:   proto.Int32(alt),
	}

	out, err := proto.Marshal(&positionPacket)
	if err != nil {
		return err
	}

	nodeNum := r.nodeNum

	radioMessage := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: &pb.MeshPacket{
				To:      nodeNum,
				WantAck: true,
				PayloadVariant: &pb.MeshPacket_Decoded{
					Decoded: &pb.Data{
						Payload:      out,
						Portnum:      pb.PortNum_POSITION_APP,
						WantResponse: true,
					},
				},
			},
		},
	}

	packet, err := proto.Marshal(&radioMessage)
	if err != nil {
		return err
	}

	if err := r.SendPacket(packet); err != nil {
		return err
	}

	return nil
}

// FactoryReset sends a factory reset command to the radio
func (r *Radio) FactoryReset() error {
	adminPacket := pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_FactoryResetDevice{
			FactoryResetDevice: 1,
		},
	}
	out, err := proto.Marshal(&adminPacket)
	if err != nil {
		return err
	}

	nodeNum := r.nodeNum

	packet, err := r.createAdminPacket(nodeNum, out)
	if err != nil {
		return err
	}

	if err := r.SendPacket(packet); err != nil {
		return err
	}

	return nil
}

// Close closes the serial port. Added so users can defer the close after opening
func (r *Radio) Close() {
	r.streamer.Close()
}
