package gomesh

import (
	"io"
	"net"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type streamer struct {
	serialPort io.ReadWriteCloser
	netPort    net.Conn
	isTCP      bool
}

func (s *streamer) Init(addr string) error {

	ip := net.ParseIP(addr)

	if ip != nil {
		tcpAddr := net.TCPAddr{IP: ip, Port: 4403}
		conn, err := net.DialTCP("tcp", nil, &tcpAddr)
		if err != nil {
			return err
		}
		s.netPort = conn
		s.isTCP = true

	} else {
		//Configure the serial port
		options := serial.OpenOptions{
			PortName:              addr,
			BaudRate:              115200,
			DataBits:              8,
			StopBits:              1,
			MinimumReadSize:       0,
			InterCharacterTimeout: 100,
			ParityMode:            serial.PARITY_NONE,
		}

		// Open the port.
		port, err := serial.Open(options)
		if err != nil {
			return err
		}

		s.serialPort = port
		s.isTCP = false

		return nil
	}

	return nil
}

func (s *streamer) Write(p []byte) error {

	if s.isTCP {
		s.netPort.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, err := s.netPort.Write(p)
		if err != nil {
			return err
		}
	} else {
		_, err := s.serialPort.Write(p)
		if err != nil {
			return err
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (s *streamer) Read(p []byte) error {

	if s.isTCP {
		s.netPort.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err := s.netPort.Read(p)
		if err != nil {
			return err
		}
	} else {
		_, err := s.serialPort.Read(p)
		if err != nil {
			return err
		}
	}

	return nil

}

func (s *streamer) Close() {
	if s.isTCP {
		s.netPort.Close()
	} else {
		s.serialPort.Close()
	}
}
