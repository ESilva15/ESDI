package telemetry

import (
	"math"
	"strconv"
	"sync"
	"time"
)

type FieldMapper struct {
	SDKKey    string
	DataType  DataType
	Transform func(any) uint64
}

// NOTE: Update the iracing SDK to write data to the same map ALWAYS, then
// I can bind that address and read directly from there on the transform
type BoundField struct {
	Key       string
	ID        FieldID
	Transform func(any, *TelemetryField)
}

var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, MaxFields*8)
		return &b
	},
}

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
// From my testing, using an uint64 bucket is around 50x faster than any/interface{}
// The IDs member is a list so that we can have multiple items on the target device
// consuming the same piece of data
// NOTE: we can optimize this via a special command that says a given piece of data
// is for multiple targets
type TelemetryField struct {
	IDs  []int16 // Identification for the serial device
	Type DataType
	Raw  uint64
	Str  string // Only to be used with DataTypeSTRING
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
func (tf *TelemetryField) Pack(dest []byte) []byte {
	// NOTE: maybe we can have a pool of these so we don't have to create them here
	// or whatever
	for _, id := range tf.IDs {
		dest = append(dest, uint8(id), uint8(id>>8))
		dest = append(dest, uint8(tf.Type))

		switch tf.Type {
		case DataTypeINT8, DataTypeUINT8, DataTypeCHAR:
			dest = append(dest, uint8(tf.Raw))
		case DataTypeINT16, DataTypeUINT16:
			dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8))
		case DataTypeINT32, DataTypeUINT32:
			dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16), uint8(tf.Raw>>24))
		case DataTypeINT64, DataTypeUINT64:
			dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16),
				uint8(tf.Raw>>24), uint8(tf.Raw>>32), uint8(tf.Raw>>40), uint8(tf.Raw>>48),
				uint8(tf.Raw>>56),
			)
		case DataTypeSTRING:
			l := min(len(tf.Str), math.MaxUint8)

			dest = append(dest, uint8(l))
			dest = append(dest, tf.Str[:l]...)
		}
	}

	return dest
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

// We use FirstField to start the count on the fields the user can select
// the first three will be for internal use
const (
	Speed FieldID = iota
	RPM
	Gear
	BrakeBias
	SessionTime
	ReplaySessionTime
	Empty
	MaxFields
)
const FirstField = Speed

var FieldNames = [MaxFields]string{
	Speed:             "Speed",
	RPM:               "RPM",
	Gear:              "Gear",
	BrakeBias:         "BrakeBias",
	SessionTime:       "SessionTime",
	ReplaySessionTime: "ReplaySessionTime",
	Empty:             "Emtpy",
}

func GetFieldName(id FieldID) string {
	if id >= MaxFields {
		return "Unknown"
	}

	return FieldNames[id]
}

var fieldNameToID map[string]FieldID

func initFieldNamesMap() {
	fieldNameToID = make(map[string]FieldID, MaxFields)
	for id, name := range FieldNames {
		fieldNameToID[name] = FieldID(id)
	}
}

func GetFieldID(name string) (FieldID, bool) {
	id, ok := fieldNameToID[name]
	return id, ok
}

// TelemetryData is
// I need to find a way of having the values be per window or some other
type TelemetryData struct {
	Values              [MaxFields]TelemetryField
	ActiveBinds         []BoundField
	InitialTime         time.Time
	PenultimateDataPoll time.Time
	LastDataPoll        time.Time
}

func NewTelemetryData() *TelemetryData {
	return &TelemetryData{}
}

func (td *TelemetryData) Pack() []byte {
	bufPtr := bufferPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]

	for _, bind := range td.ActiveBinds {
		buf = td.Values[bind.ID].Pack(buf)
	}

	// We have to copy here because we have to return the buffer
	result := make([]byte, len(buf))
	copy(result, buf)

	bufferPool.Put(&buf)

	return result
}
