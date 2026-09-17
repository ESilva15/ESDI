package uidevice

import (
	"esdi/peripheral"
	"esdi/telemetry"
)

type UIDevice struct {
	dataChan chan telemetry.TelemetryData
}

const NAME = "UIView"

func NewUIDevice() (peripheral.Peripheral, error) {
	return &UIDevice{
		dataChan: make(chan telemetry.TelemetryData, 1),
	}, nil
}

func (uid *UIDevice) SendData(data *telemetry.TelemetryData) {
	if data == nil {
		return
	}

	select {
	case uid.dataChan <- *data:
	default:
		// Drop frame if buffer is full
	}
}

func (uid *UIDevice) Name() string {
	return NAME
}

func (uid *UIDevice) DataChannel() <-chan telemetry.TelemetryData {
	return uid.dataChan
}

func (uid *UIDevice) RequiredFields() []telemetry.FieldID {
	return []telemetry.FieldID{
		telemetry.Speed,
		telemetry.Gear,
		telemetry.RPM,
	}
}
