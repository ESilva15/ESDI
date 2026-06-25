package telemetry

import "strconv"

func EmptyTransform(v any, out *TelemetryField) {
	out.Type = DataTypeCHAR
	out.Raw = uint64('-')
}

func UInt8Transform(v any, out *TelemetryField) {
	out.Type = DataTypeUINT8

	if v == nil {
		out.Raw = 0
		return
	}

	out.Raw = uint64(v.(int))
}

func FloatToStringTransform(v any, out *TelemetryField) {
	out.Type = DataTypeSTRING

	if v == nil {
		out.Str = "inv"
		return
	}

	out.Str = strconv.FormatFloat(float64(v.(float32)), 'f', 1, 32)
}

func FloatToUInt8Transform(v any, out *TelemetryField) {
	out.Type = DataTypeUINT8

	if v == nil {
		out.Raw = uint64(0)
		return
	}

	out.Raw = uint64(v.(float32))
}
