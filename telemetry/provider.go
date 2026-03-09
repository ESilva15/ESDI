// Package telemetry is our interface with our data sources
package telemetry

type TelemetryProvider interface {
	Stream() (<-chan TelemetryData, error)
	Subscribe([]FieldID)
}
