package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"esdi/providers"
	"esdi/telemetry"
	telem "esdi/telemetry"
)

var ErrNoActiveProviderAvailable = errors.New("no active provider available")

// TelemetryService will be our base struct to handle telemetry data
// It should hook to a data sink and handle it like iRacing, BeamNG, AC and so on
type TelemetryService struct {
	logger *slog.Logger
	// Streaming
	isStreaming bool
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
	// Callbacks
	startedStreaming func(ch <-chan telemetry.TelemetryData)
	// Devices data request
	// peripheralProvider func() []peripheral.Peripheral
	// getRequiredFields func() []telemetry.FieldID
}

func NewTelemetryService(
	logger *slog.Logger,
	msg chan string,
) *TelemetryService {
	newService := &TelemetryService{
		logger:      logger,
		isConnected: false,
		listeners:   make(map[string]chan telem.TelemetryData),
		Messages:    msg,
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
			slog.Info("checking if provider is still running")
			if !t.activeProvider.IsAlive(500 * time.Millisecond) {
				t.Messages <- "Healthcheck on provider failing. Dropping provider.\n"
				slog.Warn("provider healthcheck failed")
				t.dropActiveProvider()
				t.onProviderHealthCheckFailed()
				return
			}
		}
	}
}

func (t *TelemetryService) GetTelemetryProviderName() (string, error) {
	if t.activeProvider == nil {
		return "", ErrNoActiveProviderAvailable
	}

	return t.activeProvider.Name(), nil
}

// Listener Control [START] ----------------------------------------------------

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

func (t *TelemetryService) SubscribeToFields(fields []telemetry.FieldID) []string {
	t.logger.Debug("requested fields", "fields", fields)
	subscribed := t.activeProvider.Subscribe(fields)

	return subscribed
}

// Listener Control [END] ------------------------------------------------------

// Provider Control [START] ----------------------------------------------------

func (t *TelemetryService) HasActiveProvider() bool {
	if t.activeProvider == nil {
		return false
	}

	return true
}

func (t *TelemetryService) dropActiveProvider() {
	if t.cancelForward != nil {
		t.cancelForward()
	}

	t.activeProvider.StopStream()
	t.activeProvider.Close()
	t.activeProvider = nil
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
				// t.logger.Debug("checking provider: " + prov.Name)
				provider, err := prov.NewProvider(t.logger)
				if err != nil {
					// Its not running
					continue
				}

				if provider.IsAlive(500 * time.Millisecond) {
					t.onFindProvider(provider)
					return
				}
			}
		}
	}
}

// Provider Control [END] ------------------------------------------------------

// Streaming Control [START] ---------------------------------------------------

func (t *TelemetryService) StopStream() {
	t.mut.Lock()
	defer t.mut.Unlock()

	if t.cancelForward != nil {
		t.cancelForward()
		t.cancelForward = nil
	}

	t.activeProvider.StopStream()
	t.isStreaming = false
}

func (t *TelemetryService) StartStream() <-chan telemetry.TelemetryData {
	// Start the new stream
	if t.activeProvider == nil {
		slog.Debug("there's no active provider. not starting the stream")
		return nil
	}

	// Stop the provider healthcheck
	t.healthCheckCancel()

	simInCh, _ := t.activeProvider.Stream()

	// Create the context so we can control the lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	t.cancelForward = cancel

	// Multiplex this data
	go t.multiplexData(ctx, simInCh)
	t.isStreaming = true

	t.Messages <- "Provider has started streaming data\n"

	return simInCh
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
				// t.logger.Debug("sending data to listener", "listener", key, "data", data)
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

func (t *TelemetryService) IsStreaming() bool {
	return t.isStreaming
}

// Streaming Control [END] -----------------------------------------------------

// Callbacks [START] -----------------------------------------------------------

// Callbacks [END] -------------------------------------------------------------
