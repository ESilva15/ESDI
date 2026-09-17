package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"esdi/devices"
	"esdi/peripheral"
	"esdi/telemetry"
)

var ErrPeripheralAlreadyRegistered = errors.New("peripheral is already registered")

// DeviceService will handle sending the data from the telemetry service to the
// actual devices
// NOTE: create a virtual device and make it be the output window or something so
// we can just add it as a device or whatever instead of being a custom made thing
// that would be pretty cool I think
type DeviceService struct {
	Logger *slog.Logger
	// Device discovery
	mu                 sync.RWMutex
	ctxDiscovery       context.Context
	ctxDiscoveryCancel context.CancelFunc
	Devices            map[string]peripheral.Peripheral
	// Strem handling
	streamCancel context.CancelFunc
	TelemCh      <-chan telemetry.TelemetryData
	// Output
	Messages chan string
}

func NewDeviceService(logger *slog.Logger) *DeviceService {
	sharedChannel := make(chan string, 10)

	dev := &DeviceService{
		Devices:  make(map[string]peripheral.Peripheral),
		Logger:   logger,
		Messages: sharedChannel,
	}

	// Start the routine that looks for devices - should always be running in the background
	// Create a routine to poll this provider while we wait to start the stream or pause it
	dev.ctxDiscovery, dev.ctxDiscoveryCancel = context.WithCancel(context.Background())
	go dev.FindDevices()

	return dev
}

func (ds *DeviceService) FindDevices() {
	// Need to define a list of devices to search for
	// For now lets just try to find our cdashdisplay - will think about the rest later
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctxDiscovery.Done():
			// If requested to cancel we cancel background discovery
			return
		case <-ticker.C:
			for pName, peripheral := range devices.List {
				if ds.DeviceExists(pName) {
					// We already discovered this device
					continue
				}

				ds.Logger.Debug("looking for device", "name", pName)
				dev, err := peripheral.Discover()
				if err != nil {
					ds.Logger.Debug("didn't find device", "name", pName)
					continue
				}

				// Register the device we just found
				ds.RegisterDevice(dev)
			}
		}
	}
}

// func (ds *DeviceService) SubscribeFields() error {
// 	for _, dev := range ds.Devices {
// 		fields := dev.RequiredFields()
// 	}
//
// 	return nil
// }

func (ds *DeviceService) RegisterDevice(dev peripheral.Peripheral) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.DeviceExists(dev.Name()) {
		return ErrPeripheralAlreadyRegistered
	}

	ds.Devices[dev.Name()] = dev

	return nil
}

func (ds *DeviceService) GetDevice(name string) (peripheral.Peripheral, error) {
	val, ok := ds.Devices[name]
	if !ok {
		return nil, fmt.Errorf("device `%s` couldn't be found", name)
	}

	return val, nil
}

func (ds *DeviceService) DeviceExists(name string) bool {
	if _, ok := ds.Devices[name]; !ok {
		return false
	}

	return true
}

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
			ds.mu.RLock()
			for _, dev := range ds.Devices {
				dev.SendData(&data)
			}
			ds.mu.RUnlock()

			isSending.Store(false)
		}
	}
}
