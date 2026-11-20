package gomesh

import (
	"io"
	"testing"

	pb "github.com/lmatte7/meshtastic-go/pkg/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

// scriptedPort is a fake serial port that serves scripted read results: each
// chunk is returned byte by byte, and between chunks a read reports the link
// idle the way the serial driver does (zero bytes, io.EOF).
type scriptedPort struct {
	chunks  [][]byte
	written [][]byte
}

func (p *scriptedPort) Read(b []byte) (int, error) {
	for len(p.chunks) > 0 && len(p.chunks[0]) == 0 {
		p.chunks = p.chunks[1:]
		return 0, io.EOF
	}
	if len(p.chunks) == 0 {
		return 0, io.EOF
	}
	b[0] = p.chunks[0][0]
	p.chunks[0] = p.chunks[0][1:]
	return 1, nil
}

func (p *scriptedPort) Write(b []byte) (int, error) {
	p.written = append(p.written, append([]byte{}, b...))
	return len(b), nil
}

func (p *scriptedPort) Close() error { return nil }

func testRadio(chunks ...[]byte) (*Radio, *scriptedPort) {
	port := &scriptedPort{chunks: chunks}
	return &Radio{streamer: streamer{serialPort: port}}, port
}

// frameFor wraps a FromRadio message in the serial framing.
func frameFor(t *testing.T, msg *pb.FromRadio) []byte {
	t.Helper()
	payload, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte{start1, start2, byte(len(payload) >> 8), byte(len(payload))}, payload...)
}

func TestReadResponseKeepsPartialFrameAcrossCalls(t *testing.T) {
	frame := frameFor(t, &pb.FromRadio{PayloadVariant: &pb.FromRadio_ConfigCompleteId{ConfigCompleteId: 7}})
	// The link goes idle one byte before the packet is complete.
	split := len(frame) - 1
	radio, _ := testRadio(frame[:split], frame[split:])

	first, err := radio.ReadResponse(false)
	if err != nil || len(first) != 0 {
		t.Fatalf("first read = %v, %v; want no packets yet", first, err)
	}
	second, err := radio.ReadResponse(false)
	if err != nil || len(second) != 1 || second[0].GetConfigCompleteId() != 7 {
		t.Fatalf("second read = %v, %v; want the completed packet", second, err)
	}
}

func TestReadResponseSkipsNoiseBetweenFrames(t *testing.T) {
	frame := frameFor(t, &pb.FromRadio{PayloadVariant: &pb.FromRadio_ConfigCompleteId{ConfigCompleteId: 9}})
	noise := []byte("boot log text\r\n")
	radio, _ := testRadio(append(append(append([]byte{}, noise...), frame...), noise...))

	packets, err := radio.ReadResponse(false)
	if err != nil || len(packets) != 1 || packets[0].GetConfigCompleteId() != 9 {
		t.Fatalf("packets = %v, %v; want the one framed packet", packets, err)
	}
}
