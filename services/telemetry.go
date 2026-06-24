package services

import (
	"context"
	"log/slog"
	"sync"

	"esdi/providers"
	telem "esdi/telemetry"
)

// TelemetryService will be our base struct to handle telemetry data
// It should hook to a data sink and handle it like iRacing, BeamNG, AC and so on
type TelemetryService struct {
	logger *slog.Logger
	cdash  *CDashService
	// Concurrency protection
	mut           sync.RWMutex
	ativeProvider telem.TelemetryProvider
	// Channel for the UI
	listeners     map[string]chan telem.TelemetryData
	uiOutCh       chan telem.TelemetryData
	cancelForward context.CancelFunc
}

func NewTelemetryService(logger *slog.Logger, cdash *CDashService) *TelemetryService {
	newService := &TelemetryService{
		logger:    logger,
		cdash:     cdash,
		uiOutCh:   make(chan telem.TelemetryData, 100),
		listeners: make(map[string]chan telem.TelemetryData),
	}

	// Need to instantiate a default provider here
	source := "/home/esilva/Desktop/projetos/simracing_peripherals/testTelemetry/gt3_mustang_bathurst.ibt"
	firstProvider := providers.NewIRacingProvider(slog.Default(), source, "", "")
	newService.SwitchProvider(firstProvider)

	return newService
}

// func (t *TelemetryService) GetUIStream() <-chan telem.TelemetryData {
// 	return t.uiOutCh
// }

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

			select {
			case t.uiOutCh <- data:
			default:
				// Control latency
			}
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
