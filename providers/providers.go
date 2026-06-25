// Package providers
package providers

import (
	"log/slog"

	"esdi/providers/beamng"
	"esdi/providers/iracing"
	"esdi/telemetry"
)

// Make this be some kind of struct where we can access a function that returns
// the selected provider by its name
type Provider struct {
	Name     string
	Provider telemetry.TelemetryProvider
}

var Providers = map[string]Provider{
	beamng.NAME: {
		Name: beamng.NAME,
	},
	iracing.NAME: {
		Name: iracing.NAME,
	},
}

func NewIRacingProvider(logger *slog.Logger, source string,
	telemOut string, yamlOut string,
) telemetry.TelemetryProvider {
	provider, _ := iracing.NewIRacingProvider(logger, source, "", "")

	return provider
}

func NewBeamNGProvider(ip string, port int) telemetry.TelemetryProvider {
	provider, _ := beamng.NewBeamNGProvider(ip, port)
	return provider
}
