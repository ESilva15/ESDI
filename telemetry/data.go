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
// From my testing, using an uint64 bucket is around 50x faster than any
type TelemetryField struct {
	ID   int16 // Identification: I have to learn how I will use this on other devices
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
	dest = append(dest, uint8(tf.ID), uint8(tf.ID>>8))
	dest = append(dest, uint8(tf.Type))

	switch tf.Type {
	case DataTypeINT8, DataTypeUINT8:
		dest = append(dest, uint8(tf.Raw))
	case DataTypeINT16, DataTypeUINT16:
		dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8))
	case DataTypeINT32, DataTypeUINT32:
		dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16))
	case DataTypeINT64, DataTypeUINT64:
		dest = append(dest, uint8(tf.Raw), uint8(tf.Raw>>8), uint8(tf.Raw>>16),
			uint8(tf.Raw>>24), uint8(tf.Raw>>32), uint8(tf.Raw>>40), uint8(tf.Raw>>48),
			uint8(tf.Raw>>56),
		)
	case DataTypeSTRING:
		l := min(len(tf.Str), math.MaxUint8)

		dest = append(dest, uint8(l))
		dest = append(dest, tf.Str[:l]...)
	case DataTypeCHAR:
		dest = append(dest, uint8(tf.Raw))
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
	MaxFields
)
const FirstField = Speed

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

// NOTE: Replace values with a more appropriate custom field approach where
// every custom field only takes as many bytes as required
// NOTE: Add the timing fields here to count frames of data gathering and whatnot
// remember to do the same to whoever is sending data

type TelemetryData struct {
	// Values map[string]*TelemetryField
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
