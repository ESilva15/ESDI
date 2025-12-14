// Package communication will handle the data transmission with the peripheral
// device
package communication

import (
	"github.com/tarm/serial"
)

const (
	StartOfText = 0x02
	EndOfText   = 0x03
)

type CommState uint8

const (
	CommIdle CommState = iota
	CommOff
)

type Command uint8

const (
	CmdRequestID Command = iota
	CmdAckID
	CmdCreateScreen
	CmdCreateWindow
)

type DataReceiver interface {
	Read(dev *serial.Port) error
	Validate(data []byte) bool
}
