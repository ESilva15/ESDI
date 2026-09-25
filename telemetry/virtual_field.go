package telemetry

// VirtualField will derive data from telemetry primitives
// So fuel per lap predictions, compound gauge lights and so on
type VirtualField interface {
	Name() string
	Process(td *TelemetryData)
	EnsureSubscribed() []FieldID
}
