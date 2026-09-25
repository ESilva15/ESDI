// Package beamng is the BeamNG.drive data provider
package beamng

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"esdi/constants"
	"esdi/telemetry"

	bngsdk "github.com/ESilva15/gobngsdk"
)

// BeamNG is the concrete implementation of the TelemetryProvider interface
// for BeamNG.drive
// NOTE: document this please. What is a TelemetryData????
type BeamNG struct {
	logger *slog.Logger

	SDK *bngsdk.BeamNGSDK
	og  *bngsdk.Outgauge

	// data handling
	mut      sync.Mutex
	data     *telemetry.TelemetryData
	updaters [telemetry.MaxFields]func(*telemetry.TelemetryField)

	// Field Subscription management
	boundFields map[telemetry.FieldID]bool

	// stream control
	wg           sync.WaitGroup
	streamCancel context.CancelFunc

	// timing
	ticker *time.Ticker
}

const (
	NAME = constants.BeamNGProviderName
)

func NewBeamNGProvider(logger *slog.Logger, opts *bngsdk.Options) (*BeamNG, error) {
	beam, err := bngsdk.NewBngSDK(*opts)
	if err != nil {
		return &BeamNG{}, err
	}

	provider := &BeamNG{
		logger: logger.With("TelemetryProvider", NAME),
		data:   telemetry.NewTelemetryData(),
		SDK:    beam,
		og:     &bngsdk.Outgauge{},
		// Field subscription management
		boundFields: make(map[telemetry.FieldID]bool, telemetry.MaxFields),
		ticker:      time.NewTicker(time.Second / 60),
	}

	provider.updaters = [telemetry.MaxFields]func(*telemetry.TelemetryField){
		telemetry.Speed:     provider.updateSpeed,
		telemetry.Gear:      provider.updateGear,
		telemetry.RPM:       provider.updateRPM,
		telemetry.FuelLevel: provider.fuelLevel,
		// Engine Data
		telemetry.OilPress:  provider.oilPressure,
		telemetry.OilTemp:   provider.oilTemp,
		telemetry.WaterTemp: provider.engTemp,
		// Electrics (dash lights and so on)
		telemetry.PitSpeedLimiter:   provider.pitSpeedLimiter,
		telemetry.LeftIndicator:     provider.leftIndicator,
		telemetry.RightIndicator:    provider.rightIndicator,
		telemetry.Hazards:           provider.unused,
		telemetry.ABSWarningLight:   provider.absLight,
		telemetry.ParkingBrakeLight: provider.handbrakeLight,
		telemetry.TCLight:           provider.tcLight,
		telemetry.BatteryLight:      provider.batteryLight,
	}

	// Set the unset telemetry fields on the updaters as unused fields
	for k := range int(telemetry.MaxFields) {
		if provider.updaters[k] == nil {
			provider.updaters[k] = provider.unused
		}
	}

	return provider, nil
}

func (b *BeamNG) Close() {
	b.SDK.Close()
}

func (b *BeamNG) Name() string {
	return NAME
}

func (b *BeamNG) IsAlive(timeout time.Duration) bool {
	_, err := b.SDK.Update()
	return err == nil
}

func (b *BeamNG) StopStream() {
	if b.streamCancel == nil {
		return
	}

	b.streamCancel()
}

func (b *BeamNG) Stream() (<-chan telemetry.TelemetryData, error) {
	var ctx context.Context
	ctx, b.streamCancel = context.WithCancel(context.Background())

	// Start the stream
	ch := b.stream(ctx)

	return ch, nil
}

// TODO: this function is exactly the same in BeamNG drive now, and I reckon it will be the same
// In plenty other things. I should make it TelemetryData method
func (i *BeamNG) subscribe(fields []telemetry.FieldID) []string {
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

func (b *BeamNG) Subscribe(requestFields []telemetry.FieldID) []string {
	// NOTE: document how the Subscribe funtion works
	slog.Debug(fmt.Sprintf("Len Req: %d\n", len(requestFields)))

	// First we must add the virtual fields
	// we will add their dependencies and the primitives to a slice
	toBind := make([]telemetry.FieldID, telemetry.MaxFields)

	b.mut.Lock()
	for _, id := range requestFields {
		switch id {
		case telemetry.RPMStateColour:
			rpmLights := telemetry.NewRPMLights()
			b.data.VirtualBinds[rpmLights.Name()] = rpmLights
			toBind = append(toBind, rpmLights.EnsureSubscribed()...)
		case telemetry.FCCurrentLap:
			fuelCalc := telemetry.NewFuelCalculator(b.logger.WithGroup("FUEL CALC"))
			b.data.VirtualBinds[fuelCalc.Name()] = fuelCalc
			toBind = append(toBind, fuelCalc.EnsureSubscribed()...)
		default:
			// primitive telemetry field
			toBind = append(toBind, id)
		}
	}
	b.mut.Unlock()

	newSubs := b.subscribe(toBind)

	slog.Debug(fmt.Sprintf("Subscribed: %+v\n", toBind))

	return newSubs
}

// Internal

func (b *BeamNG) readData() error {
	ogSnapshot, err := b.SDK.Update()
	if err != nil {
		slog.Error("Error getting data", "error", err)
		return err
	}

	b.mut.Lock()
	b.og = ogSnapshot
	defer b.mut.Unlock()

	// Read 1 to 1 data
	for _, bind := range b.data.ActiveBinds {
		b.updaters[bind.ID](&b.data.Values[bind.ID])
	}

	// Set up virtual binds
	for _, vBind := range b.data.VirtualBinds {
		// NOTE: delete the logs here, they are really bad
		vBind.Process(b.data)
	}

	b.data.PenultimateDataPoll = b.data.LastDataPoll
	b.data.LastDataPoll = time.Now()

	return nil
}

func (b *BeamNG) stream(ctx context.Context) <-chan telemetry.TelemetryData {
	b.data.InitialTime = time.Now()
	outCh := make(chan telemetry.TelemetryData)
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		defer close(outCh)

		for {
			// Explicitly intercept cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			select {
			case <-ctx.Done():
				return
			case <-b.ticker.C:
				err := b.readData()
				if err != nil {
					if b.streamCancel != nil {
						b.streamCancel()
					}
				}

				// Publish data
				select {
				case outCh <- *b.data:
					// Successfuly read and sent data
				default:
					// skip this data, don't allow publishers to lag behind
				}
			}
		}
	}()

	return outCh
}
