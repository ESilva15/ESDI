package telemetry

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type DataType uint8

const (
	DataTypeUINT8  DataType = 0
	DataTypeINT8   DataType = 1
	DataTypeUINT16 DataType = 2
	DataTypeINT16  DataType = 3
	DataTypeSTRING DataType = 4
)

type TelemetryField struct {
	Type  DataType
	Value any
}

// Pack will pack this current TelemetryField into bytes to send over the wire
// Format:
// 0x00 - DataType
// 0x01 - if its a (u)int8
// or
// 0x01 - if its a (u)int16 - first byte
// 0x01 - if its a (u)int16 - second byte
// or
// 0x02 - str len max is 255 chars
// [0x02] - str
func (tf *TelemetryField) Pack() []byte {
	buf := new(bytes.Buffer)

	buf.WriteByte(uint8(tf.Type))

	switch tf.Type {
	case DataTypeINT8:
		buf.WriteByte(uint8(tf.Value.(int8)))
	case DataTypeUINT8:
		buf.WriteByte(uint8(tf.Value.(uint8)))
	case DataTypeINT16, DataTypeUINT16:
		binary.Write(buf, binary.LittleEndian, tf.Value)
	case DataTypeSTRING:
		str := tf.Value.(string)
		buf.WriteByte(uint8(len(str)))
		buf.WriteString(str)
	}

	return buf.Bytes()
}

func (tf *TelemetryField) String() string {
	return fmt.Sprintf("%v", tf.Value)
}

// NOTE: Replace values with a more appropriate custom field approach where
// every custom field only takes as many bytes as required
// NOTE: Add the timing fields here to count frames of data gathering and whatnot
// remember to do the same to whoever is sending data

type TelemetryData struct {
	// Values map[string]*TelemetryField
	Values map[string][32]byte
}

func NewTelemetryData() *TelemetryData {
	return &TelemetryData{
		Values: make(map[string][32]byte),
	}
}
