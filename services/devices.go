package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"esdi/devices"
	"esdi/devices/cdashdisplay"
	"esdi/telemetry"
)

// DeviceService will handle sending the data from the telemetry service to the
// actual devices
// NOTE: create a virtual device and make it be the output window or something so
// we can just add it as a device or whatever instead of being a custom made thing
// that would be pretty cool I think
type DeviceService struct {
	Logger  *slog.Logger
	Devices map[string]devices.Device
	// Strem handling
	streamCancel context.CancelFunc
	TelemCh      <-chan telemetry.TelemetryData
	// Output
	Messages chan string
}

func NewDeviceService(logger *slog.Logger) *DeviceService {
	sharedChannel := make(chan string, 10)
	return &DeviceService{
		Devices:  make(map[string]devices.Device),
		Logger:   logger,
		Messages: sharedChannel,
	}
}

func (ds *DeviceService) FindDevices() {
	// Need to define a list of devices to search for
	// For now lets just try to find our cdashdisplay - will think about the rest later

	// Find CDashDisplay
	{
		ds.Messages <- "looking for " + cdashdisplay.Name + "...\n"
		ds.Logger.Info("Looking for " + cdashdisplay.Name)

		cdashdisplay.SetLogger(ds.Logger.With("[device]", cdashdisplay.Name))

		// Create a cdashdisplay
		display, err := cdashdisplay.Discover()
		if err == nil {
			ds.Devices[cdashdisplay.Name] = display
			ds.Logger.Info("found " + cdashdisplay.Name + " on: " + display.WT.Cfg.Name)
			ds.Messages <- "found " + cdashdisplay.Name + " on: " + display.WT.Cfg.Name + "\n"
			return
		}

		ds.Logger.Info("didn't find " + cdashdisplay.Name)
		ds.Messages <- "didn't find " + cdashdisplay.Name + "\n"
		// No CDashDisplay available for one reason or another, so we don't set the
		// key
	}
}

func (ds *DeviceService) GetDevice(name string) (devices.Device, error) {
	val, ok := ds.Devices[name]
	if !ok {
		return nil, fmt.Errorf("device `%s` couldn't be found", name)
	}

	return val, nil
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
			for _, dev := range ds.Devices {
				dev.SendData(&data)
			}
			isSending.Store(false)
		}
	}
}
