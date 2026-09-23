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

	"esdi/constants"
	"esdi/telemetry"

	"github.com/ESilva15/goirsdk"
)

const (
	NAME = constants.IRacingProviderName
)

// IRacing is our iRacing telemetry data provider - its a TelemetryProvider interface
type IRacing struct {
	logger *slog.Logger
	SDK    *goirsdk.IBT

	// Data Handling
	mut      sync.Mutex
	data     *telemetry.TelemetryData
	updaters [telemetry.MaxFields]func(*telemetry.TelemetryField)
	// Field Subscription management
	boundFields map[telemetry.FieldID]bool

	// Timing information
	ticker *time.Ticker // ticker will keep polling intervals constant

	// Stream
	wg           sync.WaitGroup
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
		logger: logger,
		SDK:    sdk,
		data:   telemetry.NewTelemetryData(),
		// Field subscription management
		boundFields: make(map[telemetry.FieldID]bool, telemetry.MaxFields),
		// TODO: make this configurable from the user side
		ticker: time.NewTicker(time.Second / 60),
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

func (i *IRacing) Close() {
	// Need to find a way of gracefully closing the channel
	// close(i.streamCh)
	i.SDK.Close()
	i.ticker.Stop()
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

func (i *IRacing) stream(ctx context.Context) <-chan telemetry.TelemetryData {
	i.data.InitialTime = time.Now()
	outCh := make(chan telemetry.TelemetryData)
	i.wg.Add(1)

	go func() {
		defer i.wg.Done()
		defer close(outCh)

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
				case outCh <- *i.data:
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

	return outCh
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
	ch := i.stream(ctx)

	return ch, nil
}

func (i *IRacing) StopStream() {
	if i.streamCancel == nil {
		return
	}

	i.streamCancel()
	i.wg.Wait()
	i.streamCancel = nil
}

// TODO: this function is exactly the same in BeamNG drive now, and I reckon it will be the same
// In plenty other things. I should make it TelemetryData method
func (i *IRacing) subscribe(fields []telemetry.FieldID) []string {
	newSubscriptions := make([]string, 0, telemetry.MaxFields)

	i.mut.Lock()
	defer i.mut.Unlock()

	for _, id := range fields {
		if i.boundFields[id] {
			continue
		}

		i.data.ActiveBinds[id] = telemetry.BoundField{
			ID: id,
		}
		i.boundFields[id] = true
		newSubscriptions = append(newSubscriptions, telemetry.FieldNames[id])
	}

	// unsubscribe from fields we many not need anymore
	for key, bound := range i.boundFields {
		if _, ok := i.data.ActiveBinds[key]; bound && !ok {
			delete(i.data.ActiveBinds, key)
			i.boundFields[key] = false
		}
	}

	return newSubscriptions
}

func (i *IRacing) Subscribe(requestFields []telemetry.FieldID) []string {
	i.logger.Debug(fmt.Sprintf("Len Req: %d\n", len(requestFields)))

	// First we must add the virtual fields
	// we will add their dependencies and the primitives to a slice
	toBind := make([]telemetry.FieldID, 0, telemetry.MaxFields)

	i.mut.Lock()
	for _, id := range requestFields {
		switch id {
		case telemetry.RPMStateColour:
			rpmLights := telemetry.NewRPMLights()
			i.data.VirtualBinds[rpmLights.Name()] = rpmLights
			toBind = append(toBind, rpmLights.EnsureSubscribed()...)
		case telemetry.FCCurrentLap:
			fuelCalc := telemetry.NewFuelCalculator(i.logger.WithGroup("FUEL CALC"))
			i.data.VirtualBinds[fuelCalc.Name()] = fuelCalc
			toBind = append(toBind, fuelCalc.EnsureSubscribed()...)
		default:
			// primitive telemetry field
			toBind = append(toBind, id)
		}
	}
	i.mut.Unlock()

	newSubs := i.subscribe(toBind)

	i.logger.Debug(fmt.Sprintf("Subscribed: %+v\n", i.data.ActiveBinds))

	return newSubs
}

func (i *IRacing) Name() string {
	return NAME
}

func (i *IRacing) IsAlive(timeout time.Duration) bool {
	if !i.SDK.CheckForDataEvent(timeout) {
		return false
	}

	return true
}
