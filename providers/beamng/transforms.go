package beamng

import (
	"strconv"

	conv "esdi/conversions"

	"esdi/telemetry"
)

const (
	LapTimeFormatStr = "04:05.000"
)

func (b *BeamNG) unused(out *telemetry.TelemetryField) {
	out.Unused()
}

func (b *BeamNG) updateSpeed(out *telemetry.TelemetryField) {
	out.Type = telemetry.DataTypeUINT16
	out.Raw = uint64(conv.MsToKph(b.SDK.Data.Speed))
}

func (b *BeamNG) updateGear(out *telemetry.TelemetryField) {
	out.Type = telemetry.DataTypeSTRING
	// NOTE: stupid idea but we can cache these values
	out.Str = strconv.Itoa(int(b.SDK.Data.Gear))
}

func (b *BeamNG) updateRPM(out *telemetry.TelemetryField) {
	out.Type = telemetry.DataTypeUINT16
	out.Raw = uint64(uint16(b.SDK.Data.RPM))
}

func (b *BeamNG) fuelLevel(out *telemetry.TelemetryField) {
	telemetry.FloatToStringTransform(b.SDK.Data.Fuel, out)
}

func (b *BeamNG) oilPressure(out *telemetry.TelemetryField) {
	telemetry.FloatToStringTransform(b.SDK.Data.OilPressure, out)
}

func (b *BeamNG) oilTemp(out *telemetry.TelemetryField) {
	telemetry.FloatToStringTransform(b.SDK.Data.OilTemp, out)
}

func (b *BeamNG) engTemp(out *telemetry.TelemetryField) {
	telemetry.FloatToStringTransform(b.SDK.Data.EngTemp, out)
}

// NOTE: find how to empty this
func (b *BeamNG) pitSpeedLimiter(out *telemetry.TelemetryField) {
}

func (b *BeamNG) leftIndicator(out *telemetry.TelemetryField) {
	chr := ' '
	if b.SDK.LeftIndicator() {
		chr = '<'
	}

	out.Type = telemetry.DataTypeCHAR
	out.Raw = uint64(chr)
}

func (b *BeamNG) rightIndicator(out *telemetry.TelemetryField) {
	chr := ' '
	if b.SDK.RightIndicator() {
		chr = '>'
	}

	out.Type = telemetry.DataTypeCHAR
	out.Raw = uint64(chr)
}

func (b *BeamNG) absLight(out *telemetry.TelemetryField) {
	chr := ' '
	if b.SDK.ABS() {
		chr = 'A'
	}

	out.Type = telemetry.DataTypeCHAR
	out.Raw = uint64(chr)
}

func (b *BeamNG) handbrakeLight(out *telemetry.TelemetryField) {
	chr := ' '
	if b.SDK.Handbrake() {
		chr = 'P'
	}

	out.Type = telemetry.DataTypeCHAR
	out.Raw = uint64(chr)
}

func (b *BeamNG) tcLight(out *telemetry.TelemetryField) {
	chr := ' '
	if b.SDK.TractionControl() {
		chr = 'T'
	}

	out.Type = telemetry.DataTypeCHAR
	out.Raw = uint64(chr)
}

func (b *BeamNG) batteryLight(out *telemetry.TelemetryField) {
	chr := ' '
	if b.SDK.BatteryLight() {
		chr = 'B'
	}

	out.Type = telemetry.DataTypeCHAR
	out.Raw = uint64(chr)
}
