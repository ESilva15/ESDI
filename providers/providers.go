// Package providers
package providers

import (
	"log/slog"

	"esdi/providers/beamng"
	"esdi/providers/iracing"
	"esdi/telemetry"

	bngsdk "github.com/ESilva15/gobngsdk"
	"github.com/ESilva15/goirsdk"
)

// Make this be some kind of struct where we can access a function that returns
// the selected provider by its name
type Provider struct {
	Name        string
	NewProvider func(*slog.Logger) (telemetry.TelemetryProvider, error)
	IsRunning   func() bool // To check if this provider is up and running
}

var Providers = map[string]Provider{
	beamng.NAME: {
		Name:        beamng.NAME,
		NewProvider: NewBeamNGProvider,
		IsRunning:   beamng.IsRunning,
	},
	iracing.NAME: {
		Name:        iracing.NAME,
		NewProvider: NewLiveIRacingProvider,
		IsRunning:   iracing.IsRunning,
	},
}

func NewLiveIRacingProvider(logger *slog.Logger) (telemetry.TelemetryProvider, error) {
	provider, err := iracing.NewIRacingProvider(logger, goirsdk.Options{
		Logger:     logger,
		SourceType: goirsdk.SharedMemoryFile,
	})
	if err != nil {
		logger.Error("failed to create iRacing provider", "err", err)
		return nil, err
	}

	return provider, nil
}

func NewBeamNGProvider(logger *slog.Logger) (telemetry.TelemetryProvider, error) {
	// TODO: these should come from some kind of config
	provider, err := beamng.NewBeamNGProvider(logger, &bngsdk.Options{
		Logger:           logger.With("TelemetryProvider", beamng.NAME),
		SourceType:       bngsdk.UDPData,
		ImportUDPAddress: "127.0.0.1",
		ImportUDPPort:    4444,
	})
	if err != nil {
		logger.Error("failed to create BeamNG provider", "err", err)
		return nil, err
	}

	return provider, nil
}
