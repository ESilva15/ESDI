package telemetry

import (
	"math"
	"strconv"
	"time"
)

type DataType uint8

const (
	DataTypeUINT8  DataType = 0
	DataTypeINT8   DataType = 1
	DataTypeUINT16 DataType = 2
	DataTypeINT16  DataType = 3
	DataTypeUINT32 DataType = 4
	DataTypeINT32  DataType = 5
	DataTypeUINT64 DataType = 6
	DataTypeINT64  DataType = 7
	DataTypeSTRING DataType = 8
	DataTypeCHAR   DataType = 9
)

// TelemetryField will be the basic unit to hold telemetry data values in our
// application.
// From my testing, using an uint64 bucket is around 50x faster than any
type TelemetryField struct {
	Type DataType
	Raw  uint64
	Str  string // Only to be used with DataTypeSTRING
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
	// NOTE: maybe we can have a pool of these so we don't have to create them here
	// or whatever
	buf := make([]byte, 0, 8)

	buf = append(buf, uint8(tf.Type))

	switch tf.Type {
	case DataTypeINT8, DataTypeUINT8:
		buf = append(buf, uint8(tf.Raw))
	case DataTypeINT16, DataTypeUINT16:
		buf = append(buf, uint8(tf.Raw), uint8(tf.Raw>>8))
	case DataTypeINT32, DataTypeUINT32:
		buf = append(buf, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16))
	case DataTypeINT64, DataTypeUINT64:
		buf = append(buf, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16), uint8(tf.Raw>>32))
	case DataTypeSTRING:
		l := min(len(tf.Str), math.MaxUint8)

		buf = append(buf, uint8(l))
		buf = append(buf, tf.Str[:l]...)
	}

	return buf
}

func (tf *TelemetryField) String() string {
	switch tf.Type {
	case DataTypeSTRING:
		return tf.Str
	case DataTypeCHAR:
		return string([]byte{byte(tf.Raw)})
	case DataTypeUINT16, DataTypeUINT8:
		return strconv.FormatUint(tf.Raw, 10)
	case DataTypeINT8:
		return strconv.FormatInt(int64(int16(tf.Raw)), 10)
	case DataTypeINT16:
		return strconv.FormatInt(int64(int(tf.Raw)), 10)
	}

	return "NaN"
}

type FieldID uint16

const (
	FirstTimeStamp FieldID = iota
	PreviousTimeStamp
	LastTimeStamp
	Speed
	RPM
	Gear
	MaxFields
)

var FieldNames = [MaxFields]string{
	Speed: "Speed",
	RPM:   "RPM",
	Gear:  "Gear",
}

func GetFieldName(id FieldID) string {
	if id >= MaxFields {
		return "Uknown"
	}

	return FieldNames[id]
}

// NOTE: Replace values with a more appropriate custom field approach where
// every custom field only takes as many bytes as required
// NOTE: Add the timing fields here to count frames of data gathering and whatnot
// remember to do the same to whoever is sending data

type TelemetryData struct {
	// Values map[string]*TelemetryField
	Values              [MaxFields]TelemetryField
	InitialTime         time.Time
	PenultimateDataPoll time.Time
	LastDataPoll        time.Time
}

func NewTelemetryData() *TelemetryData {
	return &TelemetryData{}
}
