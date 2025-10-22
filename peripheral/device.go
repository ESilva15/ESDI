package peripheral

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/tarm/serial"
)

type PeripheralDevice struct {
	Port string
	Name [32]byte
}

const (
	PACKET_MAGIC = 0xDEADBEEF
	CMD_HELLO    = 1
	CMD_IDENTIFY = 2
	CMD_ACK      = 3
)

type FlarePacket struct {
	Name [16]byte
}

func parseFlare(buf []byte) *FlarePacket {
	pkt := &FlarePacket{}
	err := binary.Read(sliceReader(buf), binary.LittleEndian, pkt)
	if err != nil {
		return nil
	}
	// if pkt.Magic != PACKET_MAGIC || pkt.Cmd != CMD_HELLO {
	// 	return nil
	// }
	return pkt
}

func sliceReader(b []byte) *byteReader { return &byteReader{b: b} }

type byteReader struct{ b []byte }

func (r *byteReader) Read(p []byte) (int, error) {
	n := copy(p, r.b)
	r.b = r.b[n:]
	if n == 0 {
		time.Sleep(10 * time.Millisecond)
	}
	return n, nil
}

func (p *PeripheralDevice) Connect() error {
	cfg := &serial.Config{
		Name:        p.Port,
		Baud:        115200,
		ReadTimeout: 500 * time.Millisecond,
	}

	s, err := serial.OpenPort(cfg)
	if err != nil {
		return err
	}

	buf := make([]byte, binary.Size(FlarePacket{}))
	n, _ := s.Read(buf)
	if n == len(buf) {
		flare := parseFlare(buf)
		if flare != nil {
			s.Write([]byte{CMD_ACK})
			return nil
		}
	}

	// if we received nothing
	s.Write([]byte{CMD_IDENTIFY})

	n, _ = s.Read(buf)
	if n == len(buf) {
		pkt := parseFlare(buf)
		if pkt != nil {
			s.Write([]byte{CMD_ACK})
			return nil
		}
	}

	return fmt.Errorf("no valid device response")
}
