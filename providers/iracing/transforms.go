package iracing

import (
	"time"

	"esdi/telemetry"
)

const (
	LapTimeFormatStr = "04:05.000"
)

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

func PitSpeedLimiterTransform(v any, out *telemetry.TelemetryField) {
	out.Type = telemetry.DataTypeSTRING
	if v.(bool) {
		out.Str = "PIT"
	} else {
		out.Str = "   "
	}
}
