package telemetry

type TelemetryField struct {
	Parser func()
	Value  any
}

// NOTE: Replace values with a more appropriate custom field approach where
// every custom field only takes as many bytes as required
// NOTE: Add the timing fields here to count frames of data gathering and whatnot
// remember to do the same to whoever is sending data

type TelemetryData struct {
	// Values map[string]*TelemetryField
	Values map[string][32]byte
}

func NewTelemetryData() *TelemetryData {
	return &TelemetryData{
		Values: make(map[string][32]byte),
	}
}
