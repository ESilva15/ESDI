package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"esdi/providers"
	telem "esdi/telemetry"
)

// TelemetryService will be our base struct to handle telemetry data
// It should hook to a data sink and handle it like iRacing, BeamNG, AC and so on
type TelemetryService struct {
	logger     *slog.Logger
	devService *DeviceService
	// Concurrency protection
	mut            sync.RWMutex
	activeProvider telem.TelemetryProvider
	isConnected    bool
	// Channel for the UI
	listeners     map[string]chan telem.TelemetryData
	cancelForward context.CancelFunc
	// Output window
	Messages chan string
	// Cancel looking for providers
	CtxMonitor        context.Context
	cancelMonitor     context.CancelFunc
	CtxHealthcheck    context.Context
	healthCheckCancel context.CancelFunc
}

func NewTelemetryService(logger *slog.Logger, devServo *DeviceService) *TelemetryService {
	sharedChannel := make(chan string, 10)
	newService := &TelemetryService{
		logger:      logger,
		isConnected: false,
		devService:  devServo,
		listeners:   make(map[string]chan telem.TelemetryData),
		Messages:    sharedChannel,
	}
	newService.CtxMonitor, newService.cancelMonitor = context.WithCancel(context.Background())

	return newService
}

func (t *TelemetryService) ProviderMonitor(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.Debug("checking if provider is still running")
			if !t.activeProvider.IsAlive(500 * time.Millisecond) {
				slog.Debug("provider healthcheck failed")
				t.dropActiveProvider()
				t.onProviderHealthCheckFailed()
				return
			}
		}
	}
}

func (t *TelemetryService) onProviderHealthCheckFailed() {
	// Just restart the whole lookup process
	go t.FindProvider(t.CtxMonitor)
}

func (t *TelemetryService) onFindProvider(prov providers.Provider) {
	// Attach to the provider
	t.logger.Info("found provider for " + prov.Name)
	t.SwitchProvider(prov.NewProvider(t.logger))

	// Create a routine to poll this provider while we wait to start the stream or pause it
	t.CtxHealthcheck, t.healthCheckCancel = context.WithCancel(context.Background())
	go t.ProviderMonitor(t.CtxHealthcheck)
}

func (t *TelemetryService) onProviderStopsMidStream() {
	// clear the current provider
	// TODO: now we need to also clear the devices to restart everything,
	// if the stream stopped we have to restart the devices and everything
	t.logger.Info("cleaning dropped provider and restarting lookup service")
	t.dropActiveProvider()
	go t.FindProvider(t.CtxMonitor)
}

// TODO: add some way of retriggering this. Currently it should:
// start monitoring on startup -> find provider -> stop monitoring (when game closes for example)
func (t *TelemetryService) FindProvider(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	t.logger.Debug("monitoring for providers")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, prov := range providers.Providers {
				t.logger.Debug("checking provider: " + prov.Name)
				if prov.IsRunning() {
					t.onFindProvider(prov)
					return
				}
			}
		}
	}
}

func (t *TelemetryService) SwitchProvider(newProvider telem.TelemetryProvider) error {
	t.mut.Lock()
	defer t.mut.Unlock()

	// Clean up the current to be old provider
	if t.activeProvider != nil {
		t.dropActiveProvider()
	}

	// Assign the new provider
	t.activeProvider = newProvider

	return nil
}

func (t *TelemetryService) multiplexData(ctx context.Context, dataCh <-chan telem.TelemetryData) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-dataCh:
			if !ok {
				t.logger.Debug("something happened on the provider - stream closed")
				t.onProviderStopsMidStream()
				return
			}

			t.mut.RLock()
			for _, ch := range t.listeners {
				select {
				case ch <- data:
					// Sends data to the subscriber
				default:
					// Subscriber is full, we just skip ahead. Maybe find a how to add metrics here
				}
			}
			t.mut.RUnlock()
		}
	}
}

func (t *TelemetryService) dropActiveProvider() {
	if t.cancelForward != nil {
		t.cancelForward()
	}

	t.activeProvider.StopStream()
	t.activeProvider.Close()
	t.activeProvider = nil
}

func (t *TelemetryService) SubscribeListener(id string, bufferSize int) <-chan telem.TelemetryData {
	t.mut.Lock()
	defer t.mut.Unlock()

	// NOTE: is this truly necessary?
	// return the channel if it already exists
	if ch, exists := t.listeners[id]; exists {
		return ch
	}

	ch := make(chan telem.TelemetryData, bufferSize)
	t.listeners[id] = ch

	t.logger.Info("New stream subscriber registered", "id", id)
	return ch
}

func (t *TelemetryService) UnsubscribeListener(id string) {
	t.mut.Lock()
	defer t.mut.Unlock()

	if ch, exists := t.listeners[id]; exists {
		close(ch)
		delete(t.listeners, id)
		t.logger.Info("Stream subscriber removed", "id", id)
	}
}

func (t *TelemetryService) SubscribeToFields(fields map[int16]telem.FieldID) {
	t.activeProvider.Subscribe(fields)
}

func (t *TelemetryService) StartStream() {
	slog.Debug("Stream started")

	// Start the new stream
	if t.activeProvider == nil {
		slog.Debug("there's no active provider. not starting the stream")
		return
	}

	// Stop the provider healthcheck
	t.healthCheckCancel()

	simInCh, _ := t.activeProvider.Stream()
	// TODO: the provider needs to be able to tell the data has stopped
	// so we can restart the provider lookup routine

	// Create the context so we can control the lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	t.cancelForward = cancel

	// Multiplex this data
	go t.multiplexData(ctx, simInCh)
}

func (t *TelemetryService) StopStream() {
	t.mut.Lock()
	defer t.mut.Unlock()

	if t.cancelForward != nil {
		t.cancelForward()
		t.cancelForward = nil
	}

	t.activeProvider.StopStream()
}

func (t *TelemetryService) HasActiveProvider() bool {
	if t.activeProvider == nil {
		return false
	}

	return true
}
