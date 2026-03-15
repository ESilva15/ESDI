package iracing

import (
	telem "esdi/telemetry"
	"strconv"
	"time"
)

const (
	LapTimeFormatStr = "04:05.000"
)

func LapTimeTransform(v any, out *telem.TelemetryField) {
	lapTimeInSeconds := v.(float32)

	if lapTimeInSeconds < 0 {
		lapTimeInSeconds = 0
	}

	wholeSeconds := int64(lapTimeInSeconds)
	lapTime := time.Unix(wholeSeconds, int64((lapTimeInSeconds-float32(wholeSeconds))*1e9))

	out.Type = telem.DataTypeSTRING
	out.Str = lapTime.Format(LapTimeFormatStr)
}

func EmptyTransform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeCHAR
	out.Raw = uint64('-')
}

func PitSpeedLimiterTransform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeSTRING
	if v.(bool) {
		out.Str = "PIT"
	} else {
		out.Str = "   "
	}
}

func UInt8Transform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeUINT8
	out.Raw = uint64(v.(int))
}

func FloatToStringTransform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeSTRING
	out.Str = strconv.FormatFloat(float64(v.(float32)), 'f', 1, 32)
}

func FloatToUInt8Transform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeUINT8
	out.Raw = uint64(v.(float32))
}
