package telemetry

import "strconv"

func EmptyTransform(v any, out *TelemetryField) {
	out.Type = DataTypeCHAR
	out.Raw = uint64('-')
}

func UInt8Transform(v int, out *TelemetryField) {
	out.Type = DataTypeUINT8

	if v < 0 {
		out.Raw = uint64(0)
		return
	}

	out.Raw = uint64(v)
}

func FloatToStringTransform(v float32, out *TelemetryField) {
	out.Type = DataTypeSTRING

	if v < 0.0 {
		out.Str = "inv"
		return
	}

	out.Str = strconv.FormatFloat(float64(v), 'f', 1, 32)
}

func FloatToUInt8Transform(v float32, out *TelemetryField) {
	out.Type = DataTypeUINT8

	if v < 0.0 {
		out.Raw = uint64(0)
		return
	}

	out.Raw = uint64(v)
}
