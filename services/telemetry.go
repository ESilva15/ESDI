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
	mut           sync.RWMutex
	ativeProvider telem.TelemetryProvider
	isConnected   bool
	// Channel for the UI
	listeners     map[string]chan telem.TelemetryData
	cancelForward context.CancelFunc
	// Output window
	Messages chan string
	// Cancel looking for providers
	CtxMonitor    context.Context
	cancelMonitor context.CancelFunc
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

func (t *TelemetryService) OnFindProvider(prov providers.Provider) {
	t.SwitchProvider(prov.NewProvider(t.logger))
	t.logger.Debug("Found provider for " + prov.Name)
}

// TODO: add some way of retriggering this. Currently it should:
// start monitoring on startup -> find provider -> stop monitoring
func (t *TelemetryService) FindProvider(ctx context.Context, callback func(providers.Provider),
) {
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
					callback(prov)
					return
				}
				t.logger.Debug("  wasn't read")
			}
		}
	}
}

func (t *TelemetryService) SwitchProvider(newProvider telem.TelemetryProvider) error {
	t.mut.Lock()
	defer t.mut.Unlock()

	// Clean up the current to be old provider
	if t.ativeProvider != nil {
		t.dropActiveProvider()
	}

	// Assign the new provider
	t.ativeProvider = newProvider

	return nil
}

func (t *TelemetryService) multiplexData(ctx context.Context, dataCh <-chan telem.TelemetryData) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-dataCh:
			if !ok {
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

	t.ativeProvider.StopStream()
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
	t.ativeProvider.Subscribe(fields)
}

func (t *TelemetryService) StartStream() {
	slog.Debug("Stream started")

	// Start the new stream
	simInCh, _ := t.ativeProvider.Stream()
	// if err != nil {
	// 	// NOTE
	// }

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

	t.ativeProvider.StopStream()
}
