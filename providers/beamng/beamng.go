// Package beamng is the BeamNG.drive data provider
package beamng

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"esdi/telemetry"

	bngsdk "github.com/ESilva15/gobngsdk"
)

// BeamNG is the concrete implementation of the TelemetryProvider interface
// for BeamNG.drive
// NOTE: document this please. What is a TelemetryData????
type BeamNG struct {
	SDK *bngsdk.BeamNGSDK

	// data handling
	mut      sync.Mutex
	data     *telemetry.TelemetryData
	updaters [telemetry.MaxFields]func(*telemetry.TelemetryField)

	// stream control
	streamCh     chan telemetry.TelemetryData
	streamCancel context.CancelFunc

	// timing
	ticker *time.Ticker
}

const (
	NAME = "BeamNG.drive"
)

func NewBeamNGProvider(ip string, port int) (*BeamNG, error) {
	beam, err := bngsdk.Init(ip, port)
	if err != nil {
		return &BeamNG{}, err
	}

	provider := &BeamNG{
		streamCh: make(chan telemetry.TelemetryData, 1),
		data:     telemetry.NewTelemetryData(),
		SDK:      &beam,
		ticker:   time.NewTicker(time.Second / 60),
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
		// Adjustements
		telemetry.BrakeBias:       provider.unused,
		telemetry.ABSSetting:      provider.unused,
		telemetry.TCSetting:       provider.unused,
		telemetry.ThrottleSetting: provider.unused,
		// Lap Data
		telemetry.LapLastLapTime: provider.unused,
		telemetry.LapNumber:      provider.unused,
		// Tire Data
		telemetry.LFtempL: provider.unused,
		telemetry.LFtempM: provider.unused,
		telemetry.LFtempR: provider.unused,
		telemetry.RFtempL: provider.unused,
		telemetry.RFtempM: provider.unused,
		telemetry.RFtempR: provider.unused,
		telemetry.LRtempL: provider.unused,
		telemetry.LRtempM: provider.unused,
		telemetry.LRtempR: provider.unused,
		telemetry.RRtempL: provider.unused,
		telemetry.RRtempM: provider.unused,
		telemetry.RRtempR: provider.unused,
		// Session Data
		telemetry.SessionTime:       provider.unused,
		telemetry.ReplaySessionTime: provider.unused,
		telemetry.Empty:             provider.unused,
		// Virtual Fields -- fields derived from primitive fields
		// RPM Dash Lights
		telemetry.RPMStateColour: provider.unused,
		// Fuel Calculator
		telemetry.FCCurrentLap:   provider.unused,
		telemetry.FCLastLap:      provider.unused,
		telemetry.FCAverage:      provider.unused,
		telemetry.FCExpectedLaps: provider.unused,
	}

	for k := range int(telemetry.MaxFields) {
		// One by one so you won't forget it next time
		if provider.updaters[k] == nil {
			msg := fmt.Sprintf("BeamNG updaters aren't all assigned. Currently unassgined: %s",
				telemetry.FieldNames[k])
			panic(msg)
		}
	}

	return provider, nil
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
	b.stream(ctx)

	return b.streamCh, nil
}

func (b *BeamNG) Subscribe(requestFields map[int16]telemetry.FieldID) {
	// NOTE: document how the Subscribe funtion works
	slog.Debug(fmt.Sprintf("Len Req: %d\n", len(requestFields)))

	b.data.ActiveBinds = make([]telemetry.BoundField, 0, len(requestFields))

	// First we must add the virtual fields
	// we will add their dependencies and the primitives to a slice
	pendingBinds := make([]telemetry.FieldID, telemetry.MaxFields)

	for winID, id := range requestFields {
		b.data.Values[id].IDs = append(b.data.Values[id].IDs, winID)

		switch id {
		case telemetry.RPMStateColour:
			b.data.VirtualBinds = append(b.data.VirtualBinds, telemetry.NewRPMLights())
		case telemetry.FCCurrentLap:
			b.data.VirtualBinds = append(b.data.VirtualBinds,
				telemetry.NewFuelCalculator(slog.Default().WithGroup("FUEL CALC")))
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

		b.data.ActiveBinds = append(b.data.ActiveBinds, binding)
		boundCheck[id] = true
	}

	slog.Debug(fmt.Sprintf("Subscribed: %+v\n", b.data.ActiveBinds))
}

// Internal

func (b *BeamNG) readData() {
	slog.Debug("READING THIS DATA")
	// BUG: getting stuck in here
	err := b.SDK.ReadData()
	slog.Debug("THE DATA WAS READ")
	if err != nil {
		slog.Error("Error getting data", "error", err)
		return
	}

	b.mut.Lock()
	defer b.mut.Unlock()

	// Read 1 to 1 data
	slog.Debug("Reading normal data binds")
	for _, bind := range b.data.ActiveBinds {
		slog.Debug("Current bind: ", "id", bind.ID)
		b.updaters[bind.ID](&b.data.Values[bind.ID])
	}

	// Set up virtual binds
	slog.Debug("Entering virtual binds loop")
	for _, vBind := range b.data.VirtualBinds {
		// NOTE: delete the logs here, they are really bad
		slog.Debug("Processing virtual binds")
		vBind.Process(b.data)
	}

	b.data.PenultimateDataPoll = b.data.LastDataPoll
	b.data.LastDataPoll = time.Now()
}

func (b *BeamNG) stream(ctx context.Context) {
	b.data.InitialTime = time.Now()

	go func() {
		for {
			// Explicitly intercept cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			// NOTE: add a method to check if there's data available, or make this happen

			select {
			case <-ctx.Done():
				return
			case <-b.ticker.C:
				slog.Debug("READING DATA")
				b.readData()
				slog.Debug("READ DATA")

				// Publish data
				select {
				case b.streamCh <- *b.data:
					slog.Debug("PUBLISHED DATA")
				default:
					// skip this data, don't allow publishers to lag behind
				}
			}
		}
	}()
}
