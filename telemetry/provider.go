// Package telemetry is our interface with our data sources
package telemetry

type TelemetryProvider interface {
	StopStream()
	Stream() (<-chan TelemetryData, error)
	Subscribe(map[int16]FieldID)
	Stalled() bool // Return true if no fresh data is coming
}
