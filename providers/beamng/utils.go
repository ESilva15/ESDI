package beamng

import (
	"log/slog"
	"net"
	"time"
)

func IsRunning() bool {
	// The address should be loaded from some type of configuration
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4443")
	if err != nil {
		return false
	}

	conn, err := net.ListenUDP("udp", addr)
	defer conn.Close()
	if err != nil {
		slog.Debug("failed to listen to udp socket: " + err.Error())
		return false
	}

	buf := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))

	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return false
	}

	// Think of a better number or something
	return n >= 80
}

// Stalled
// TODO: needs to be implemented
func (i *BeamNG) Stalled() bool {
	return false
}
