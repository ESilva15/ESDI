package peripheral

import (
	"time"

	"github.com/tarm/serial"
)

type PeripheralDeviceState int

const (
	StateDiscoveredStr = "discovered"
	StateConnectedStr  = "connected"
	StateUnknownStr    = "unknown"
)

const (
	StateUnknown PeripheralDeviceState = iota
	StateDiscovered
	StateConnected
)

func (s PeripheralDeviceState) String() string {
	switch s {
	case StateDiscovered:
		return StateDiscoveredStr
	case StateConnected:
		return StateConnectedStr
	case StateUnknown:
		fallthrough
	default:
		return StateUnknownStr
	}
}

type PeripheralDevice struct {
	State     PeripheralDeviceState
	CommState CommState
	Name      [32]byte
	WT        *WalkieTalkie
}

func NewPeripheralDevice(port string) *PeripheralDevice {
	return &PeripheralDevice{
		State:     StateUnknown,
		CommState: CommOff,
		WT: &WalkieTalkie{
			Cfg: &serial.Config{
				Name:        port,
				Baud:        115200,
				ReadTimeout: 500 * time.Millisecond,
			},
		},
	}
}

func (p *PeripheralDevice) Probe() error {
	err := p.WT.TurnOn()
	if err != nil {
		return nil
	}

	// Papers, please!
	_, err = p.WT.RequestIdentification()
	if err != nil {
		return err
	}

	// Verify the papers
	var pkt IdentificationPacket
	err = pkt.Read(p.WT)
	if err != nil {
		// Throw the man in the gulag!
		return err
	}

	// Acknowledge successful identification
	_, err = p.WT.AknowledgeIdentification()
	if err != nil {
		return err
	}

	// copy the data
	p.Merge(&pkt)
	p.ToConnectedIdling()

	return nil
}

func (p *PeripheralDevice) ToConnectedIdling() {
	p.State = StateConnected
	p.CommState = CommIdle
}

func (p *PeripheralDevice) Merge(packet *IdentificationPacket) {
	p.Name = packet.Name
}

func (p *PeripheralDevice) Disconnect() error {
	return p.WT.TurnOff()
}
