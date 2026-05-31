package telemetry

type RPMLights struct {
	State string
}

func NewRPMLights() *RPMLights {
	return &RPMLights{
		State: "WHITE",
	}
}

func (rl *RPMLights) Process(td *TelemetryData) {
	curRPM := td.Values[RPM].Raw

	td.Values[RPMStateColour].Type = DataTypeSTRING
	if curRPM >= 2000 && curRPM < 4000 {
		td.Values[RPMStateColour].Str = "GREEN"
	} else if curRPM >= 4000 && curRPM < 6000 {
		td.Values[RPMStateColour].Str = "YELLOW"
	} else if curRPM >= 6000 {
		td.Values[RPMStateColour].Str = "RED"
	} else {
		td.Values[RPMStateColour].Str = "WHITE"
	}
}
