package telemetry

import (
	"strconv"
	"sync"
	"time"
)

func init() {
	fieldNameToID = make(map[string]FieldID, MaxFields)
	for id, name := range FieldNames {
		fieldNameToID[name] = FieldID(id)
	}
}

// NOTE: allow the user to create custom data things. For example, iRacing provides
// multiple surface temps, but I guess the user doesn't want all of them at once.
// allow him to make something that allows some data transformation to occur.

type FieldMapper struct {
	SDKKey    string
	DataType  DataType
	Transform func(any) uint64
}

// VirtualField will derive data from telemetry primitives
// So fuel per lap predictions, compound gauge lights and so on
type VirtualField interface {
	Process(td *TelemetryData)
	EnsureSubscribed() []FieldID
}

// NOTE: Update the iracing SDK to write data to the same map ALWAYS, then
// I can bind that address and read directly from there on the transform

// BoundField is the data structure we use to bind the telemetry provider's data
// to our internal telemetry fields
type BoundField struct {
	ID     FieldID
	Update func(out *TelemetryField)
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
	// IDs  []int16 // Identification for the serial device
	Type DataType
	Raw  uint64
	Str  string // Only to be used with DataTypeSTRING
}

func (tf *TelemetryField) Unused() {
	tf.Type = DataTypeCHAR
	tf.Raw = uint64('-')
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

type FieldID = uint16

// We use FirstField to start the count on the fields the user can select
// the first three will be for internal use
const (
	Speed FieldID = iota
	RPM
	Gear
	FuelLevel
	// Engine Data
	OilPress
	OilTemp
	WaterTemp
	// Engine Warnings
	PitSpeedLimiter
	// Electrics (dash lights and so on)
	LeftIndicator
	RightIndicator
	Hazards
	ABSWarningLight
	ParkingBrakeLight
	TCLight
	BatteryLight
	// Adjustements
	BrakeBias
	ABSSetting
	TCSetting
	ThrottleSetting
	// Lap Data
	LapLastLapTime
	LapNumber
	// Tire Data
	LFtempL
	LFtempM
	LFtempR
	RFtempL
	RFtempM
	RFtempR
	LRtempL
	LRtempM
	LRtempR
	RRtempL
	RRtempM
	RRtempR
	// Session Data
	SessionTime
	ReplaySessionTime
	Empty
	// Virtual Fields -- fields derived from primitive fields
	// RPM Dash Lights
	RPMStateColour
	// Fuel Calculator
	FCCurrentLap
	FCLastLap
	FCAverage
	FCExpectedLaps
	MaxFields
)
const FirstField = Speed

var FieldNames = [MaxFields]string{
	Speed:     "Speed",
	RPM:       "RPM",
	Gear:      "Gear",
	FuelLevel: "Fuel Level",
	// Engine Data
	OilPress:  "Oil Pressure",
	OilTemp:   "Oil Temperature",
	WaterTemp: "Water Temperature",
	// Engine Warnings
	PitSpeedLimiter: "Pit Speed Limiter",
	// Electrics (dash lights and so on)
	LeftIndicator:     "Left Indicator",
	RightIndicator:    "Right Indicator",
	Hazards:           "Hazards",
	ABSWarningLight:   "ABS Dash Light",
	ParkingBrakeLight: "Parking Brake Dash Light",
	TCLight:           "Traction Control Light",
	BatteryLight:      "Battery Light",
	// Ajustments
	BrakeBias:       "BrakeBias",
	ABSSetting:      "ABS Control",
	TCSetting:       "TC Control",
	ThrottleSetting: "Throttle Control",
	// Lap Data
	LapLastLapTime: "Last Lap Time",
	LapNumber:      "Lap Number",
	// Tire Data
	LFtempL: "LF Surface Temp Left",
	LFtempM: "LF Surface Temp Mid",
	LFtempR: "LF Surface Temp Right",
	RFtempL: "RF Surface Temp Left",
	RFtempM: "RF Surface Temp Mid",
	RFtempR: "RF Surface Temp Right",
	LRtempL: "LR Surface Temp Left",
	LRtempM: "LR Surface Temp Mid",
	LRtempR: "LR Surface Temp Right",
	RRtempL: "RR Surface Temp Left",
	RRtempM: "RR Surface Temp Mid",
	RRtempR: "RR Surface Temp Right",
	// Session Data
	SessionTime:       "SessionTime",
	ReplaySessionTime: "ReplaySessionTime",
	// Virtual Fields
	// RPM Dash Lights
	RPMStateColour: "RPM State Colour",
	// Fuel Calculator
	FCCurrentLap:   "Fuel Current Lap",
	FCLastLap:      "Fuel Last Lap",
	FCAverage:      "Fuel Average Usage",
	FCExpectedLaps: "Fuel Expected Laps",
	Empty:          "Emtpy",
}

func GetFieldName(id FieldID) string {
	if id >= MaxFields {
		return "Unknown"
	}

	return FieldNames[id]
}

var fieldNameToID map[string]FieldID

func GetFieldID(name string) (FieldID, bool) {
	id, ok := fieldNameToID[name]
	return id, ok
}

// TelemetryData is
// I need to find a way of having the values be per window or some other
type TelemetryData struct {
	Values              [MaxFields]TelemetryField
	ActiveBinds         []BoundField
	VirtualBinds        []VirtualField
	InitialTime         time.Time
	PenultimateDataPoll time.Time
	LastDataPoll        time.Time
}

func NewTelemetryData() *TelemetryData {
	return &TelemetryData{}
}
