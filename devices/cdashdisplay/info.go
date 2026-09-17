package cdashdisplay

import (
	"esdi/peripheral/devices"
	"esdi/telemetry"
)

// TODO: I believe we don't need the #esdi/peripheral/devices thing anymore
const (
	ID   = devices.CDashDisplayDevID
	NAME = devices.CDashDisplayDevName
)

func (cds *CDashDisplay) Name() string {
	return NAME
}

func (cds *CDashDisplay) RequiredFields() map[int16]telemetry.FieldID {
	fields := make(map[int16]telemetry.FieldID, len(cds.State.Layout.Windows))

	for _, w := range cds.State.Layout.Windows {
		fieldID, _ := telemetry.GetFieldID(w.UIData.TelemetryField)
		fields[w.UIData.IDX] = fieldID
	}

	return fields
}
