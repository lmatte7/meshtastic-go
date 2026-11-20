package gomesh

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	pb "github.com/lmatte7/meshtastic-go/pkg/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

const start1 = byte(0x94)
const start2 = byte(0xc3)
const headerLen = 4
const maxToFromRadioSize = 512
const broadcastNum = 0xffffffff

// configRequestID is sent with want_config_id; the radio echoes it back in
// config_complete_id once it has finished sending its configuration and node
// database.
const configRequestID = 42

// configTimeout bounds how long GetRadioInfo waits for the radio to finish
// answering a configuration request. ESP32 boards reboot when the serial port
// is opened and need several seconds before they respond at all. It is a
// variable so tests can shorten it.
var configTimeout = 30 * time.Second

// configRetryInterval is how often the configuration request is repeated
// while the radio has not answered yet.
const configRetryInterval = 3 * time.Second

// Radio holds the port and serial io.ReadWriteCloser struct to maintain one serial connection
type Radio struct {
	streamer streamer
	nodeNum  uint32
	// frame holds the bytes of a packet that has started arriving but is not
	// complete yet, so an idle timeout in the middle of a packet does not
	// lose it.
	frame []byte
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

// ReadResponse reads from the radio until one FromRadio packet is complete,
// which it returns, or until the link goes idle, in which case it returns no
// packets. Callers loop; returning each packet as it completes keeps a busy
// link (such as TCP with packets arriving faster than the read deadline)
// from holding up output, commands, cancellation and timeouts.
//
// Each packet arrives framed as [START1, START2, length (2 bytes, big endian),
// protobuf]. Bytes outside a frame (for example boot or debug text) are
// skipped until the next START1/START2 pair. A packet that is only partly
// received when the link goes idle is kept and completed on the next call.
// The timeout argument is kept for compatibility and ignored. The result is a
// slice so existing callers keep working; it holds at most one packet.
func (r *Radio) ReadResponse(timeout bool) ([]*pb.FromRadio, error) {
	b := make([]byte, 1)

	for {
		n, err := r.streamer.Read(b)
		if n > 0 {
			if packet := r.feedFrame(b[0]); packet != nil {
				return []*pb.FromRadio{packet}, nil
			}
		}

		switch {
		case err == nil && n > 0:
			continue
		case err == nil, errors.Is(err, os.ErrDeadlineExceeded):
			// Idle: a zero-byte read, or the TCP read deadline expired.
			return nil, nil
		case errors.Is(err, io.EOF) && !r.streamer.isTCP:
			// Idle: the serial driver reports the inter-character timeout
			// as a zero-byte read, which os.File turns into io.EOF.
			return nil, nil
		default:
			// On TCP, io.EOF means the radio closed the connection; any
			// other error on either transport means the link is gone.
			return nil, err
		}
	}
}

// feedFrame adds one received byte to the frame being assembled and returns
// the decoded packet when it completes the frame, or nil.
func (r *Radio) feedFrame(c byte) *pb.FromRadio {
	r.frame = append(r.frame, c)

	switch {
	case len(r.frame) == 1 && r.frame[0] != start1:
		r.frame = r.frame[:0]
	case len(r.frame) == 2 && r.frame[1] != start2:
		r.frame = r.frame[:0]
	case len(r.frame) == headerLen && frameLength(r.frame) > maxToFromRadioSize:
		// Not a real header; resynchronise on the next start byte.
		r.frame = r.frame[:0]
	case len(r.frame) > headerLen && len(r.frame) == headerLen+frameLength(r.frame):
		packet := &pb.FromRadio{}
		err := proto.Unmarshal(r.frame[headerLen:], packet)
		r.frame = r.frame[:0]
		if err == nil {
			return packet
		}
	}
	return nil
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
	if r.nodeNum == 0 {
		return errors.New("radio configuration did not include its node number")
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

	// Traffic arrived but the radio never confirmed the dump complete, so
	// the configuration (and possibly MyInfo) may be missing. Callers must
	// not act on it.
	return nil, fmt.Errorf("radio did not finish sending its configuration within %s (%d packets received)", configTimeout, len(responses))
}

// SendTextMessage sends a free form text message to other radios
func (r *Radio) SendTextMessage(message string, to int64, channel int64) error {
	_, err := r.SendTextMessageWithID(message, to, channel)
	return err
}

// SendTextMessageWithID sends a free form text message to other radios and
// returns the packet ID assigned to it. The radio reports the fate of the
// message in a QueueStatus with this mesh_packet_id and in a ROUTING_APP
// packet whose request_id matches it (an acknowledgement or an error).
func (r *Radio) SendTextMessageWithID(message string, to int64, channel int64) (uint32, error) {
	var address int64
	if to == 0 {
		address = broadcastNum
	} else {
		address = to
	}

	if len(message) > int(pb.Constants_DATA_PAYLOAD_LEN) {
		return 0, errors.New("message too large")
	}

	// A random non-zero 32-bit id, as the firmware expects.
	packetID := rand.Uint32()
	for packetID == 0 {
		packetID = rand.Uint32()
	}

	radioMessage := pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: &pb.MeshPacket{
				To:      uint32(address),
				WantAck: true,
				Id:      packetID,
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
		return 0, err
	}

	if err := r.SendPacket(out); err != nil {
		return 0, err
	}

	return packetID, nil
}

// SetRadioOwner sets the owner of the radio visible on the public mesh
func (r *Radio) SetRadioOwner(name string) error {
	return r.SetRadioOwnerNames(name, "")
}

// SetRadioOwnerNames sets the radio's long and short name. The short name is
// at most four characters; when empty it is derived from the long name.
func (r *Radio) SetRadioOwnerNames(name, short string) error {
	if len(name) <= 2 {
		return errors.New("name too short")
	}
	if short == "" {
		short = name
		if len(short) > 4 {
			short = short[:4]
		}
	}
	if len(short) > 4 {
		return errors.New("short name must be at most 4 characters")
	}

	adminPacket := pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetOwner{
			SetOwner: &pb.User{
				LongName:  name,
				ShortName: short,
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
	"lf": pb.Config_LoRaConfig_LONG_FAST,
	"lm": pb.Config_LoRaConfig_LONG_MODERATE,
	"lt": pb.Config_LoRaConfig_LONG_TURBO,
	"ms": pb.Config_LoRaConfig_MEDIUM_SLOW,
	"mf": pb.Config_LoRaConfig_MEDIUM_FAST,
	"mt": pb.Config_LoRaConfig_MEDIUM_TURBO,
	"sl": pb.Config_LoRaConfig_SHORT_SLOW,
	"sf": pb.Config_LoRaConfig_SHORT_FAST,
	"st": pb.Config_LoRaConfig_SHORT_TURBO,
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
