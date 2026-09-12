// Package iracing is in the providers and provides telemety to our desktop
// application. It needs to establish a relationship between the desktop data
// structure of our app and the data iRacing provides
package iracing

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"esdi/telemetry"

	"github.com/ESilva15/goirsdk"
)

const (
	NAME = "iRacing"
)

// IRacing is our iRacing telemetry data provider - its a TelemetryProvider interface
type IRacing struct {
	logger *slog.Logger
	SDK    *goirsdk.IBT

	// Data Handling
	mut      sync.Mutex
	data     *telemetry.TelemetryData
	updaters [telemetry.MaxFields]func(*telemetry.TelemetryField)

	// Timing information
	ticker *time.Ticker // ticker will keep polling intervals constant

	// Stream
	streamCh     chan telemetry.TelemetryData
	streamCancel context.CancelFunc
}

func NewIRacingProvider(
	logger *slog.Logger,
	opts goirsdk.Options,
) (*IRacing, error) {
	var err error

	sdk, err := goirsdk.Init(opts)
	if err != nil {
		logger.Error("failed to open the IRSDK instance")
		return &IRacing{}, err
	}

	provider := &IRacing{
		logger:   logger,
		SDK:      sdk,
		data:     telemetry.NewTelemetryData(),
		streamCh: make(chan telemetry.TelemetryData, 1),
		// NOTE: This is because I stupidly recorded a test IBT file in 240
		// TODO: make this configurable from the user side
		ticker: time.NewTicker(time.Second / 240),
	}

	provider.updaters = [telemetry.MaxFields]func(*telemetry.TelemetryField){
		telemetry.Speed:     provider.speed,
		telemetry.Gear:      provider.gear,
		telemetry.RPM:       provider.rpm,
		telemetry.FuelLevel: provider.fuelLevel,
		// Engine Data
		telemetry.OilPress:  provider.oilPress,
		telemetry.OilTemp:   provider.oilTemp,
		telemetry.WaterTemp: provider.waterTemp,
		// EngineWarnings
		telemetry.PitSpeedLimiter: provider.pitSpeedLimiter,
		// Adjustements
		telemetry.BrakeBias:       provider.brakeBias,
		telemetry.ABSSetting:      provider.absSetting,
		telemetry.TCSetting:       provider.tcSetting,
		telemetry.ThrottleSetting: provider.throttleSetting,
		// Lap Data
		telemetry.LapLastLapTime: provider.lapTime,
		telemetry.LapNumber:      provider.lapNumber,
		// case telemetry.LFtempM:
		// 	binding.Transform = func(v any, out *telemetry.TelemetryField) {
		// 		out.Type = telemetry.DataTypeSTRING
		// 		out.Str = strconv.FormatFloat(float64(v.(float32)), 'f', 1, 32)
		// 	}
		// case telemetry.SessionTime:
		// 	binding.Transform = func(v any, out *telemetry.TelemetryField) {
		// 		out.Type = telemetry.DataTypeSTRING
		// 		out.Str = strconv.FormatFloat(v.(float64), 'f', 1, 32)
		// 	}
		// case telemetry.ReplaySessionTime:
		// 	binding.Transform = func(v any, out *telemetry.TelemetryField) {
		// 		out.Type = telemetry.DataTypeSTRING
		// 		out.Str = strconv.FormatFloat(v.(float64), 'f', 1, 32)
		// 	}
		// case telemetry.Empty:
		// 	binding.Transform = telemetry.EmptyTransform
		// }
	}

	// Set the unset telemetry fields on the updaters as unused fields
	for k := range int(telemetry.MaxFields) {
		if provider.updaters[k] == nil {
			provider.updaters[k] = provider.unused
		}
	}

	return provider, nil
}

func (i *IRacing) isDataAvailable() bool {
	// Its offline telemetry, data must be available
	if i.SDK.File == nil {
		return false
	}

	// Check if live telemetry is on
	if !i.SDK.IsConnected() {
		return false
	}

	return true
}

func (i *IRacing) stream(ctx context.Context) {
	i.data.InitialTime = time.Now()

	go func() {
		// Put this into the configuration file
		consecutiveTimeouts := 0
		maxTimeouts := 30
		dataEvTimeout := 100
		for {
			// Explicitly intercept cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			if i.SDK.CheckForDataEvent(time.Duration(dataEvTimeout) * time.Millisecond) {
				consecutiveTimeouts = 0
				i.readData()

				// Publish data
				select {
				case i.streamCh <- *i.data:
				default:
					// skip this data, don't allow publishers to lag behind
				}
			} else {
				consecutiveTimeouts++
				if consecutiveTimeouts >= maxTimeouts {
					i.logger.Info("Telemetry stream stalled")
					if i.streamCancel != nil {
						i.streamCancel()
					}
					return
				}
			}
		}
	}()
}

func (i *IRacing) readData() {
	i.mut.Lock()
	defer i.mut.Unlock()

	var err error

	_, err = i.SDK.Update(time.Millisecond * 16)
	if err != nil {
		return
	}

	// Read 1 to 1 data using our updaters
	for _, bind := range i.data.ActiveBinds {
		i.updaters[bind.ID](&i.data.Values[bind.ID])
	}

	// Set up virtual binds
	for _, vBind := range i.data.VirtualBinds {
		vBind.Process(i.data)
	}

	i.data.PenultimateDataPoll = i.data.LastDataPoll
	i.data.LastDataPoll = time.Now()
}

// Telemetry Provider Interface

// Stream returns a channel that we will use to funnel the telemetry data back to the
// UI, which then should broadcast it to the devices
func (i *IRacing) Stream() (<-chan telemetry.TelemetryData, error) {
	var ctx context.Context
	ctx, i.streamCancel = context.WithCancel(context.Background())

	// Start the stream
	i.stream(ctx)

	return i.streamCh, nil
}

func (i *IRacing) StopStream() {
	if i.streamCancel == nil {
		return
	}

	i.streamCancel()
	i.streamCancel = nil
}

func (i *IRacing) Subscribe(requestFields map[int16]telemetry.FieldID) {
	i.logger.Debug(fmt.Sprintf("Len Req: %d\n", len(requestFields)))

	i.data.ActiveBinds = make([]telemetry.BoundField, 0, len(requestFields))

	// First we must add the virtual fields
	// we will add their dependencies and the primitives to a slice
	pendingBinds := make([]telemetry.FieldID, 0, telemetry.MaxFields)

	for winID, id := range requestFields {
		i.data.Values[id].IDs = append(i.data.Values[id].IDs, winID)

		switch id {
		case telemetry.RPMStateColour:
			i.data.VirtualBinds = append(i.data.VirtualBinds, telemetry.NewRPMLights())
		case telemetry.FCCurrentLap:
			i.data.VirtualBinds = append(i.data.VirtualBinds,
				telemetry.NewFuelCalculator(i.logger.WithGroup("FUEL CALC")))
		default:
			// primitive telemetry field
			pendingBinds = append(pendingBinds, id)
		}
	}

	boundCheck := make(map[telemetry.FieldID]bool)

	// Now that we know all the fields we need to bind we follow the binding procedure
	for _, id := range pendingBinds {
		// Check if we already bound this FieldID
		if boundCheck[id] {
			continue
		}

		binding := telemetry.BoundField{
			ID: id,
		}

		i.data.ActiveBinds = append(i.data.ActiveBinds, binding)
		boundCheck[id] = true
	}

	i.logger.Debug(fmt.Sprintf("Subscribed: %+v\n", i.data.ActiveBinds))
}
