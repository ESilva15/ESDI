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
	cdash          *CDashService
	ActiveProvider telemetry.TelemetryProvider
}

func NewTelemetryService(logger *slog.Logger, cdash *CDashService) *TelemetryService {
	return &TelemetryService{
		logger: logger,
		cdash:  cdash,
	}
}

func (t *TelemetryService) setIRacingProvider() {
	path := "/home/esilva/Desktop/projetos/simracing_peripherals/testTelemetry/gt3_mustang_bathurst.ibt"
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

	// Somewhere around here we have to tell the device service to also send data
	t.cdash.StreamData(stream)

	return stream
}
