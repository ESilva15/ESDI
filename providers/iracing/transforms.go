package iracing

import (
	"time"

	conv "esdi/conversions"
	"esdi/telemetry"

	"github.com/ESilva15/goirsdk"
)

const (
	LapTimeFormatStr = "04:05.000"
)

func getVar[T any](vars map[string]goirsdk.Var, key string, fallback T) T {
	raw, ok := vars[key]
	if !ok {
		return fallback
	}

	val, ok := raw.Value.(T)
	if !ok {
		return fallback
	}

	return val
}

func (i *IRacing) unused(out *telemetry.TelemetryField) {
	out.Unused()
}

func (i *IRacing) speed(out *telemetry.TelemetryField) {
	speed := getVar(i.SDK.Vars.Vars, "Speed", float32(0))

	out.Type = telemetry.DataTypeUINT16
	out.Raw = uint64(conv.MsToKph(speed))
}

func (i *IRacing) gear(out *telemetry.TelemetryField) {
	gear := getVar(i.SDK.Vars.Vars, "Gear", int32(0))

	out.Type = telemetry.DataTypeCHAR

	switch {
	case gear == 0:
		out.Raw = uint64('N')
	case gear < 0:
		out.Raw = uint64('R')
	case gear > 0 && gear < 10:
		out.Raw = uint64('0' + gear)
	default:
		out.Raw = uint64('?') // Fallback
	}
}

func (i *IRacing) rpm(out *telemetry.TelemetryField) {
	rpm := getVar(i.SDK.Vars.Vars, "RPM", float32(0))

	out.Type = telemetry.DataTypeUINT16
	out.Raw = uint64(uint16(rpm))
}

func (i *IRacing) fuelLevel(out *telemetry.TelemetryField) {
	fuelLevel := getVar(i.SDK.Vars.Vars, "FuelLevel", float32(-1.0))
	telemetry.FloatToStringTransform(fuelLevel, out)
}

// EngineWarnings
func (i *IRacing) pitSpeedLimiter(out *telemetry.TelemetryField) {
	out.Type = telemetry.DataTypeSTRING
	out.Str = "   "

	if i.SDK.PitSpeedLimiter() {
		out.Str = "PIT"
	}
}

// EngineData ↓
func (i *IRacing) oilPress(out *telemetry.TelemetryField) {
	oilPress := getVar(i.SDK.Vars.Vars, "OilPress", float32(-1.0))
	telemetry.FloatToStringTransform(oilPress, out)
}

func (i *IRacing) oilTemp(out *telemetry.TelemetryField) {
	oilTemp := getVar(i.SDK.Vars.Vars, "OilTemp", float32(-1.0))
	telemetry.FloatToStringTransform(oilTemp, out)
}

func (i *IRacing) waterTemp(out *telemetry.TelemetryField) {
	waterTemp := getVar(i.SDK.Vars.Vars, "WaterTemp", float32(-1.0))
	telemetry.FloatToStringTransform(waterTemp, out)
}

// EngineData ↑

// Adjustments ↓
func (i *IRacing) brakeBias(out *telemetry.TelemetryField) {
	bb := getVar(i.SDK.Vars.Vars, "dcBrakeBias", float32(-1.0))
	telemetry.FloatToStringTransform(bb, out)
}

func (i *IRacing) absSetting(out *telemetry.TelemetryField) {
	abs := getVar(i.SDK.Vars.Vars, "dcABS", float32(-1.0))
	telemetry.FloatToUInt8Transform(abs, out)
}

func (i *IRacing) tcSetting(out *telemetry.TelemetryField) {
	tc := getVar(i.SDK.Vars.Vars, "dcTractionControl", float32(-1.0))
	telemetry.FloatToUInt8Transform(tc, out)
}

func (i *IRacing) throttleSetting(out *telemetry.TelemetryField) {
	throttle := getVar(i.SDK.Vars.Vars, "dcThrottleShape", float32(-1.0))
	telemetry.FloatToUInt8Transform(throttle, out)
}

// Adjustments ↑

// Laps ↓
func (i *IRacing) lapTime(out *telemetry.TelemetryField) {
	lapTimeInSeconds := getVar(i.SDK.Vars.Vars, "LapLastLapTime", float32(0))

	if lapTimeInSeconds < 0 {
		lapTimeInSeconds = 0
	}

	wholeSeconds := int64(lapTimeInSeconds)
	lapTime := time.Unix(wholeSeconds, int64((lapTimeInSeconds-float32(wholeSeconds))*1e9))

	out.Type = telemetry.DataTypeSTRING
	out.Str = lapTime.Format(LapTimeFormatStr)
}

func (i *IRacing) lapNumber(out *telemetry.TelemetryField) {
	lapNumber := getVar(i.SDK.Vars.Vars, "Lap", int(0))
	telemetry.UInt8Transform(lapNumber, out)
}

// Laps ↑
