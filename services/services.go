// Package services interacts with the other libraries required for this UI
package services

import (
	"errors"
	"log/slog"
)

type Orchestrator struct {
	DeviceService    *DeviceService
	TelemetryService *TelemetryService
	Messages         chan string
}

func NewOrchestrator(logger *slog.Logger) (*Orchestrator, error) {
	msg := make(chan string, 10)

	devService := NewDeviceService(logger.With("service", "DeviceService"), msg)

	telemService := NewTelemetryService(logger.With("service", "TelemetryService"), msg)
	if telemService == nil {
		return nil, errors.New("failed to create telemetry service")
	}

	go telemService.FindProvider(telemService.CtxMonitor)

	// Setup device service callbacks
	devService.telemetryProvider = telemService.GetTelemetryProviderName
	devService.triggerFieldSubscription = telemService.SubscribeToAllFields

	// Setup telemetry service callbacks
	telemService.getRequiredFields = devService.GetRequiredFields

	return &Orchestrator{
		DeviceService:    devService,
		TelemetryService: telemService,
		Messages:         msg,
	}, nil
}
