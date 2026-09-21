package services

import (
	"context"
	"fmt"

	telem "esdi/telemetry"
)

func (t *TelemetryService) onProviderHealthCheckFailed() {
	// Just restart the whole lookup process
	go t.FindProvider(t.CtxMonitor)
}

func (t *TelemetryService) onFindProvider(prov telem.TelemetryProvider) {
	// Attach to the provider
	t.logger.Info("found provider for " + prov.Name())
	t.Messages <- fmt.Sprintf("Found provider \"%s\"\n", prov.Name())
	err := t.SwitchProvider(prov)
	if err != nil {
		t.Messages <- fmt.Sprintf("Failed to switch to provider: %+v\n", err.Error())
		t.logger.Error("failed to switch to provider onFindProvider", "err", err)
		return
	}

	// Start the healthcheck on our provider so we can drop it if it stops
	t.CtxHealthcheck, t.healthCheckCancel = context.WithCancel(context.Background())
	go t.ProviderMonitor(t.CtxHealthcheck)

	// Tell the devices service we got a provider
	t.OnProviderFound(prov.Name())
}

func (t *TelemetryService) onProviderStopsMidStream() {
	// clear the current provider
	// TODO: now we need to also clear the devices to restart everything,
	// if the stream stopped we have to restart the devices and everything
	t.Messages <- "Telemetry provider stopped mid stream\n"
	t.logger.Info("cleaning dropped provider and restarting lookup service")
	t.dropActiveProvider()
	go t.FindProvider(t.CtxMonitor)
}
