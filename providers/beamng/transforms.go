package beamng

import (
	"strconv"
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
	out.Type = telemetry.DataTypeSTRING

	// NOTE: relook at this. Just copied it here but smells bad
	gear := 0
	if val, ok := v.(int32); ok {
		gear = int(val)
	} else if val, ok := v.(int); ok {
		gear = val
	}

	// NOTE: stupid idea but we can cache these values
	out.Str = strconv.Itoa(gear)
}
