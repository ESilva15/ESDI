package iracing

import (
	telem "esdi/telemetry"
	"strconv"
)

func PitSpeedLimiterTransform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeSTRING
	if v.(bool) {
		out.Str = "PIT"
	} else {
		out.Str = "   "
	}
}

func FloatToStringTransform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeSTRING
	out.Str = strconv.FormatFloat(float64(v.(float32)), 'f', 1, 32)
}

func FloatToUInt8Transform(v any, out *telem.TelemetryField) {
	out.Type = telem.DataTypeUINT8
	out.Raw = uint64(v.(float32))
}
