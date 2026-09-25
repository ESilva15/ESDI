package cdashdisplay

import (
	"esdi/peripheral/communication/packets"
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

func (cds *CDashDisplay) HealthCheck() bool {
	// Send the command
	var health packets.HealthCheck
	err := cds.WT.SendCommand(healthCheckCMDID, []byte{0x01, 0x02, 0x03, 0x04}, &health)
	if err != nil {
		return false
	}

	return true
}
