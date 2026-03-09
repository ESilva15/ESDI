package services

import (
	providerir "esdi/providers/iracing"
	"log/slog"

	telemetry "esdi/telemetry"
)

// TelemetryService will be our base struct to handle telemetry data
// It should hook to a data sink and handle it like iRacing, BeamNG, AC and so on
type TelemetryService struct {
	logger         *slog.Logger
	ActiveProvider telemetry.TelemetryProvider
}

func NewTelemetryService(logger *slog.Logger) *TelemetryService {
	return &TelemetryService{
		logger: logger,
	}
}

func (t *TelemetryService) setIRacingProvider() {
	path := "/home/esilva/Desktop/projetos/simracing_peripherals/testTelemetry/supercars_indianapolis.ibt"
	provider, _ := providerir.NewIRacingProvider(t.logger, path, "", "")

	t.ActiveProvider = provider
}

func (t *TelemetryService) SetProvider(provider string) *TelemetryService {
	if t.ActiveProvider != nil {
		// Gotta do something here to clean up before switching
	}

	switch provider {
	case "iRacing":
		// Set up iRacing
		t.setIRacingProvider()
	default:
		return nil
	}

	return t
}

func (t *TelemetryService) StartStream() <-chan telemetry.TelemetryData {
	stream, _ := t.ActiveProvider.Stream()
	return stream
}
