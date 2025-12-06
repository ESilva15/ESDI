// Package devices will define the available devices we can communicate with
package devices

import "esdi/peripheral/types"

// IDs for our devices. They need to be correctly mapped on the devices
// themselves so we can discover them
const (
	CDashDisplayDevID = 0x01
	ESBtnBoxDevID     = 0x02
)

// DeviceMap maps the implemented devices
// When adding a new device we append it here too
var DeviceMap = map[int]Device{
	CDashDisplayDevID: CDashDisplay,
	ESBtnBoxDevID:     ESBtnBox,
}

type Device struct {
	ID   uint8
	Name [32]byte
	API  map[string]DeviceCMD
}

type DeviceCMDHeader struct {
	Payload [64]byte
}

type DeviceCMDPayload struct {
	Payload [64]byte
}

// DeviceCMDFn defines the basic type for the functions the devices can have
// The function will receive a []byte that should be the arguments the user
// types in the REPL - or however this will be used
type DeviceCMDFn func(data []byte)

type DeviceCMD struct {
	Identifier types.CommandID
	Name       string
	Desc       string
	Header     DeviceCMDHeader
	Data       DeviceCMDPayload
	Fn         DeviceCMDFn
}
