// Package providers
package providers

import (
	"log/slog"

	"esdi/providers/beamng"
	"esdi/providers/iracing"
	"esdi/telemetry"

	"github.com/ESilva15/goirsdk"
)

// Make this be some kind of struct where we can access a function that returns
// the selected provider by its name
type Provider struct {
	Name        string
	NewProvider func(*slog.Logger) telemetry.TelemetryProvider
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

func NewLiveIRacingProvider(logger *slog.Logger) telemetry.TelemetryProvider {
	provider, _ := iracing.NewIRacingProvider(logger, goirsdk.Options{
		Logger:     logger,
		SourceType: goirsdk.SharedMemoryFile,
	})

	return provider
}

func NewBeamNGProvider(logger *slog.Logger) telemetry.TelemetryProvider {
	// TODO: Get from some kind of config or whatever
	provider, _ := beamng.NewBeamNGProvider("127.0.0.1", 4444, logger)

	return provider
}
