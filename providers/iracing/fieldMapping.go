package iracing

import (
	"esdi/telemetry"
)

type FieldMapper struct {
	SDKKey    string
	DataType  telemetry.DataType
	Transform func(any) uint64
}

// NOTE: Update the iracing SDK to write data to the same map ALWAYS, then
// I can bind that address and read directly from there on the transform
type boundField struct {
	Key       string
	ID        telemetry.FieldID
	Transform func(any, *telemetry.TelemetryField)
}

var internalToSDKFieldNames = map[telemetry.FieldID]string{
	telemetry.Speed: "Speed",
	telemetry.Gear:  "Gear",
	telemetry.RPM:   "RPM",
}
