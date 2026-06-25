package beamng

import (
	"time"

	"esdi/telemetry"
)

const (
	LapTimeFormatStr = "04:05.000"
)

// NOTE: this is a copy from iracing provider. Change it however necessary
func LapTimeTransform(v any, out *telemetry.TelemetryField) {
	lapTimeInSeconds := v.(float32)

	if lapTimeInSeconds < 0 {
		lapTimeInSeconds = 0
	}

	wholeSeconds := int64(lapTimeInSeconds)
	lapTime := time.Unix(wholeSeconds, int64((lapTimeInSeconds-float32(wholeSeconds))*1e9))

	out.Type = telemetry.DataTypeSTRING
	out.Str = lapTime.Format(LapTimeFormatStr)
}

// NOTE: this is a copy from iracing provider. Change this however necessary in the future
// for beamng.
func PitSpeedLimiterTransform(v any, out *telemetry.TelemetryField) {
	out.Type = telemetry.DataTypeSTRING
	if v.(bool) {
		out.Str = "PIT"
	} else {
		out.Str = "   "
	}
}

func GearTransform(v any, out *telemetry.TelemetryField) {
	out.Type = telemetry.DataTypeCHAR

	gear := 0
	if val, ok := v.(int32); ok {
		gear = int(val)
	} else if val, ok := v.(int); ok {
		gear = val
	}

	switch {
	case gear == 0:
		out.Raw = uint64('N') // ASCII 78
	case gear < 0:
		out.Raw = uint64('R') // ASCII 82
	case gear > 0 && gear < 10:
		// Quickest way to turn 1 into '1', 2 into '2', etc.
		// ASCII '0' is 48, so 48 + 1 = 49 ('1')
		out.Raw = uint64('0' + gear)
	default:
		out.Raw = uint64('?') // Fallback
	}
}
