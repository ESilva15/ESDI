package cdashdisplay

import (
	"math"

	"esdi/telemetry"
)

// In this file we will place all structs that are 1:1 representation of the
// types in the device transport layer ->

const (
	ShowIDFalse uint8 = 0
	ShowIDTrue  uint8 = 1
)

const (
	WinTypeBASE uint8 = iota
	WinTypeBAR
	WinTypeSTRING
	WinTypeTABLE
)

var WinTYPES = []string{"BASE", "BAR", "STRING", "TABLE"}

type UIDimensions struct {
	X0     uint16 `yaml:"X0"`
	Y0     uint16 `yaml:"Y0"`
	Width  uint16 `yaml:"Width"`
	Height uint16 `yaml:"Height"`
}

type UIDecorations struct {
	BGColour     uint16 `yaml:"BGColour"`
	FGColour     uint16 `yaml:"FGColour"`
	TitleColour  uint16 `yaml:"TitleColour"`
	BorderColour uint16 `yaml:"BorderColour"`
	TitleSize    uint8  `yaml:"TitleSize"`
	TextSize     uint8  `yaml:"TextSize"`
	HasBorder    uint8  `yaml:"HasBorder"`
	Padding      uint8  `yaml:"Padding"`
}

// NOTE: we can't use FString32 for this - too many bytes
// NOTE: use a bit flags for this options instead
type UIWindowOpts struct {
	ShowID       uint8     `yaml:"ShowID"`
	WinType      uint8     `yaml:"WinType"`
	PreviewValue FString32 `yaml:"PreviewValue"`
}

type UIWindow struct {
	Dims  UIDimensions  `yaml:"Dims"`
	Decor UIDecorations `yaml:"Decor"`
	Opts  UIWindowOpts  `yaml:"Opts"`
	Title FString32     `yaml:"Title"`
}

type DesktopUIWindow struct {
	UIWindow
	UIData DesktopUIData
}

type DesktopUIData struct {
	IDX            int16  `yaml:"WID"`
	TelemetryField string `yaml:"TelemetryField"`
}

type UIWindowUpdatePacket struct {
	WinID  int16
	Window UIWindow
}

// TODO: this is here because this is the only device I use like this
// once I add another serial device I will move this somewhere else that can
// be reused by all devices but still is decoupled from the telemetry package
func (cds *CDashDisplay) encodePacket(td *telemetry.TelemetryData) []byte {
	bufPtr := cds.bufPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]

	for fieldID, windowIDs := range cds.fieldToWindows {
		if int(fieldID) >= len(td.Values) || len(windowIDs) == 0 {
			continue
		}

		tf := &td.Values[fieldID]

		for _, winID := range windowIDs {
			buf = packField(winID, tf, buf)
		}
	}

	// We have to copy here because we have to return the buffer
	result := make([]byte, len(buf))
	copy(result, buf)

	cds.bufPool.Put(&buf)

	return result
}

// Pack will pack this current TelemetryField into bytes to send over the wire
// Format:
// 0x00 - Field ID
// 0x00 |
// 0x01 - DataType
// 0x02 - if its a (u)int8
// or
// 0x02 - if its a (u)int16 - first byte
// 0x02 - if its a (u)int16 - second byte
// or
// 0x02 - str len max is 255 chars
// [0x02] - str
func packField(winID int16, tf *telemetry.TelemetryField, dest []byte) []byte {
	// NOTE: maybe we can have a pool of these so we don't have to create them here
	// or whatever
	dest = append(dest, uint8(winID), uint8(winID>>8))
	dest = append(dest, uint8(tf.Type))

	switch tf.Type {
	case telemetry.DataTypeINT8, telemetry.DataTypeUINT8, telemetry.DataTypeCHAR:
		dest = append(dest, uint8(tf.Raw))
	case telemetry.DataTypeINT16, telemetry.DataTypeUINT16:
		dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8))
	case telemetry.DataTypeINT32, telemetry.DataTypeUINT32:
		dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16), uint8(tf.Raw>>24))
	case telemetry.DataTypeINT64, telemetry.DataTypeUINT64:
		dest = append(
			dest, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16),
			uint8(tf.Raw>>24), uint8(tf.Raw>>32), uint8(tf.Raw>>40), uint8(tf.Raw>>48),
			uint8(tf.Raw>>56),
		)
	case telemetry.DataTypeSTRING:
		l := min(len(tf.Str), math.MaxUint8)

		dest = append(dest, uint8(l))
		dest = append(dest, tf.Str[:l]...)
	}

	return dest
}
