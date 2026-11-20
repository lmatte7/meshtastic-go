package gomesh

import (
	"io"
	"net"
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

func TestReadResponseReportsClosedTCPConnection(t *testing.T) {
	client, server := net.Pipe()
	radio := &Radio{streamer: streamer{netPort: client, isTCP: true}}

	// A packet followed by the radio closing the connection: the packet is
	// delivered first, then the closed link is reported as an error rather
	// than as the link being idle.
	frame := frameFor(t, &pb.FromRadio{PayloadVariant: &pb.FromRadio_ConfigCompleteId{ConfigCompleteId: 3}})
	go func() {
		server.Write(frame)
		server.Close()
	}()

	packets, err := radio.ReadResponse(false)
	if err != nil || len(packets) != 1 || packets[0].GetConfigCompleteId() != 3 {
		t.Fatalf("first read = %v, %v; want the packet", packets, err)
	}
	if _, err := radio.ReadResponse(false); err == nil {
		t.Fatal("a closed TCP connection should be reported as an error")
	}
}

func TestReadResponseYieldsEachPacketPromptly(t *testing.T) {
	first := frameFor(t, &pb.FromRadio{PayloadVariant: &pb.FromRadio_ConfigCompleteId{ConfigCompleteId: 1}})
	second := frameFor(t, &pb.FromRadio{PayloadVariant: &pb.FromRadio_ConfigCompleteId{ConfigCompleteId: 2}})
	// Both frames arrive back to back with no idle gap between them.
	radio, _ := testRadio(append(append([]byte{}, first...), second...))

	for want := uint32(1); want <= 2; want++ {
		packets, err := radio.ReadResponse(false)
		if err != nil || len(packets) != 1 || packets[0].GetConfigCompleteId() != want {
			t.Fatalf("read %d = %v, %v; want exactly packet %d", want, packets, err, want)
		}
	}
	if packets, _ := radio.ReadResponse(false); len(packets) != 0 {
		t.Errorf("idle link should yield no packets, got %v", packets)
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
