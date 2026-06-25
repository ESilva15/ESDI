// Package beamng is the BeamNG.drive data provider
package beamng

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	conv "esdi/conversions"
	"esdi/telemetry"

	bngsdk "github.com/ESilva15/gobngsdk"
)

type BeamNG struct {
	SDK *bngsdk.BNGSDK

	// data handling
	mut  sync.Mutex
	data *telemetry.TelemetryData

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

	return &BeamNG{
		streamCh: make(chan telemetry.TelemetryData, 1),
		data:     telemetry.NewTelemetryData(),
		SDK:      &beam,
		ticker:   time.NewTicker(time.Second / 60),
	}, nil
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

		// Translate the UI FieldIDs to this provider's field names
		sdkKey, ok := internalToSDKFieldNames[id]
		if !ok {
			slog.Debug("failed to get internal id")
			// Need to find a way to pass a message saying something wasn't right
			continue
		}

		binding := telemetry.BoundField{
			Key: sdkKey,
			ID:  id,
		}

		switch id {
		case telemetry.Speed:
			binding.Transform = func(v any, out *telemetry.TelemetryField) {
				out.Type = telemetry.DataTypeUINT16
				out.Raw = uint64(conv.MsToKph(v.(float32)))
			}
		case telemetry.Gear:
			binding.Transform = GearTransform
		case telemetry.RPM:
			binding.Transform = func(v any, out *telemetry.TelemetryField) {
				out.Type = telemetry.DataTypeUINT16
				out.Raw = uint64(uint16(v.(float32)))
			}
		case telemetry.FuelLevel:
			binding.Transform = telemetry.FloatToStringTransform
		// Engine Data
		case telemetry.OilPress:
			binding.Transform = telemetry.FloatToStringTransform
		case telemetry.OilTemp:
			binding.Transform = telemetry.FloatToStringTransform
		case telemetry.WaterTemp:
			binding.Transform = telemetry.FloatToStringTransform
		// Something else
		case telemetry.PitSpeedLimiter:
			binding.Transform = PitSpeedLimiterTransform
		// Adjustements
		case telemetry.BrakeBias:
			binding.Transform = telemetry.FloatToStringTransform
		case telemetry.ABSSetting:
			binding.Transform = telemetry.FloatToUInt8Transform
		case telemetry.TCSetting:
			binding.Transform = telemetry.FloatToUInt8Transform
		case telemetry.ThrottleSetting:
			binding.Transform = telemetry.FloatToUInt8Transform
		case telemetry.LFtempM:
			binding.Transform = func(v any, out *telemetry.TelemetryField) {
				out.Type = telemetry.DataTypeSTRING
				out.Str = strconv.FormatFloat(float64(v.(float32)), 'f', 1, 32)
			}
		case telemetry.SessionTime:
			binding.Transform = func(v any, out *telemetry.TelemetryField) {
				out.Type = telemetry.DataTypeSTRING
				out.Str = strconv.FormatFloat(v.(float64), 'f', 1, 32)
			}
		case telemetry.ReplaySessionTime:
			binding.Transform = func(v any, out *telemetry.TelemetryField) {
				out.Type = telemetry.DataTypeSTRING
				out.Str = strconv.FormatFloat(v.(float64), 'f', 1, 32)
			}
		case telemetry.Empty:
			binding.Transform = telemetry.EmptyTransform
		case telemetry.LapLastLapTime:
			binding.Transform = LapTimeTransform
		case telemetry.LapNumber:
			binding.Transform = telemetry.UInt8Transform
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
		v := b.SDK.DataDict[bind.Key]

		bind.Transform(v, &b.data.Values[bind.ID])
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

			// We start by checking if we do or do not have data available
			// if !b.isDataAvailable() {
			// 	continue
			// }

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
