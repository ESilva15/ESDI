package peripheral

import (
	comm "esdi/peripheral/communication"
	pack "esdi/peripheral/communication/packets"
	"esdi/peripheral/devices"
	"esdi/peripheral/types"
	"fmt"
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
	CommState comm.CommState
	ID        uint8
	Name      string
	WT        *comm.WalkieTalkie
	DeviceAPI *devices.Device
}

func NewPeripheralDevice(port string) *PeripheralDevice {
	return &PeripheralDevice{
		State:     StateUnknown,
		CommState: comm.CommOff,
		WT: &comm.WalkieTalkie{
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
		return err
	}

	// Send the identification command
	cmd := comm.CmdRequestID
	var response pack.IdentificationPacket
	err = p.WT.SendCommand(cmd, []byte{0x06, 0x07, 0x08, 0x09}, &response)
	if err != nil {
		return err
	}

	// copy the data
	p.Merge(&response)
	p.ToConnectedIdling()

	return nil
}

func (p *PeripheralDevice) SendCommand(cmd types.Command, payload []byte) error {
	fmt.Println("Send command:", cmd)
	fmt.Println("With payload:", payload)

	var ack pack.AckPacket
	err := p.WT.SendCommand(cmd, payload, &ack)
	if err != nil {
		return err
	}

	if ack.AckByte == 0x06 {
		fmt.Println("Succesfully received ack")
	}

	return nil
}

func (p *PeripheralDevice) ToConnectedIdling() {
	p.State = StateConnected
	p.CommState = comm.CommIdle
}

func (p *PeripheralDevice) Merge(packet *pack.IdentificationPacket) {
	p.Name = string(packet.Name[:])
	p.ID = packet.DeviceID
}

func (p *PeripheralDevice) Disconnect() error {
	return p.WT.TurnOff()
}
