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

func (cds *CDashDisplay) RequiredFields() []telemetry.FieldID {
	fields := make([]telemetry.FieldID, 0, len(cds.State.Layout.Windows))

	for _, w := range cds.State.Layout.Windows {
		if fieldID, ok := telemetry.GetFieldID(w.UIData.TelemetryField); ok {
			fields = append(fields, fieldID)
		}
	}

	return fields
}
