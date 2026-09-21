package services

import (
	"context"
	"log/slog"
	"sync/atomic"

	"esdi/devices"
	"esdi/peripheral"
	"esdi/telemetry"
)

// DeviceService is the API for the peripherals
type DeviceService struct {
	Logger *slog.Logger
	// Device discovery
	PSS                *PeripheralStateStore // Store to track peripheral state
	ctxDiscovery       context.Context
	ctxDiscoveryCancel context.CancelFunc
	// Strem handling
	streamCancel context.CancelFunc
	TelemCh      <-chan telemetry.TelemetryData
	// Output
	Messages chan string
	// Callbacks
	// Telemetry service data fetchers
	telemetryProvider func() (string, error)
}

func NewDeviceService(logger *slog.Logger, msg chan string) *DeviceService {
	dev := &DeviceService{
		PSS: NewPeripheralStateStore(
			logger.With("Service", "PeripheralStateStore"), devices.List, msg,
		),
		Logger:   logger,
		Messages: msg,
	}

	// Start the routine that looks for devices - should always be running in the background
	// Create a routine to poll this provider while we wait to start the stream or pause it
	dev.ctxDiscovery, dev.ctxDiscoveryCancel = context.WithCancel(context.Background())
	go dev.FindDevices()

	// Set the callbacks for PSS
	dev.PSS.telemetryProvider = dev.getTelemetryProvider

	return dev
}

// Getters [START] -------------------------------------------------------------
// This function is currently only being used by PSS, we may have to find a better
// pattern for this
func (ds *DeviceService) getTelemetryProvider() (string, error) {
	return ds.telemetryProvider()
}

func (ds *DeviceService) GetDevices() []peripheral.Peripheral {
	snapshot := ds.PSS.GetStates()
	peripherals := make([]peripheral.Peripheral, 0, len(snapshot))

	for _, state := range snapshot {
		if state.State != DeviceIsConnected {
			continue
		}
		peripherals = append(peripherals, state.Peripheral)
	}

	return peripherals
}

func (ds *DeviceService) GetPeripheral(pname string) (peripheral.Peripheral, error) {
	return ds.PSS.GetPeripheral(pname)
}

func (ds *DeviceService) PeripheralExists(pname string) bool {
	_, err := ds.PSS.GetPeripheral(pname)
	return err == nil
}

// Getters [END] ---------------------------------------------------------------

// Actions [START] -------------------------------------------------------------
func (ds *DeviceService) StartStream() {
	// NOTE: i'm using this pattern a whole lot. Maybe I can create a struct to handle this
	var ctx context.Context
	ctx, ds.streamCancel = context.WithCancel(context.Background())

	go ds.transmit(ctx)
}

func (ds *DeviceService) StopStream() {
	if ds.streamCancel == nil {
		return
	}

	ds.streamCancel()
	ds.streamCancel = nil
}

// SetTelemetryChannel sets the TelemCh to the passed channel
func (ds *DeviceService) SetTelemetryChannel(ch <-chan telemetry.TelemetryData) {
	ds.TelemCh = ch
}

// Actions [END] ---------------------------------------------------------------

// transmit will send the data to the devices themselves
func (ds *DeviceService) transmit(ctx context.Context) {
	var isSending atomic.Bool

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-ds.TelemCh:
			if !ok {
				return
			}

			if isSending.Load() {
				continue
			}

			isSending.Store(true)

			// TODO: make a copy of the data and send that copy instead of keeping
			// the data locked
			for _, dev := range ds.PSS.GetStates() {
				if dev.State != DeviceIsConnected {
					continue
				}

				err := dev.Peripheral.SendData(&data)

				if err == peripheral.ErrDeviceTimedOut {
					ds.onDeviceTimedOut(dev.device.Name)
				}
			}

			isSending.Store(false)
		}
	}
}

// Callbacks [START] -----------------------------------------------------------

// Callbacks [END] -------------------------------------------------------------
