// Package communication will handle the data transmission with the peripheral
package communication

import (
	"esdi/peripheral/types"
)

type CommState uint8

const (
	CommIdle CommState = iota
	CommOff
)

const (
	CmdRequestID types.Command = iota
	CmdAckID
	CmdCreateScreen
	CmdCreateWindow
)
