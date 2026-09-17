// Package telemetry is our interface with our data sources
package telemetry

import "time"

type TelemetryProvider interface {
	StopStream()
	Stream() (<-chan TelemetryData, error)
	Subscribe([]FieldID)
	IsAlive(time.Duration) bool
	Name() string
	Close()
}
