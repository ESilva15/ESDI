// Package iracing is in the providers and provides telemety to our desktop
// application. It needs to establish a relationship between the desktop data
// structure of our app and the data iRacing provides
package iracing

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	conv "esdi/conversions"
	telem "esdi/telemetry"

	"github.com/ESilva15/goirsdk"
)

// IRacing is our iRacing telemetry data provider - its a TelemetryProvider interface
type IRacing struct {
	logger *slog.Logger
	SDK    *goirsdk.IBT

	// Data Handling
	mut  sync.Mutex
	data *telem.TelemetryData

	// Timing information
	ticker *time.Ticker // ticker will keep polling intervals constant

	// Stream
	streamCh     chan telem.TelemetryData
	streamCancel context.CancelFunc
}

func NewIRacingProvider(
	logger *slog.Logger,
	source string,
	telemOut string,
	yamlOut string,
) (*IRacing, error) {
	var err error

	// Open the input file if provided
	var file *os.File = nil
	if source != "" {
		file, err = os.Open(source)
		if err != nil {
			log.Fatalf("Failed to open IBT file: %v", err)
		}
	}

	sdk, err := goirsdk.Init(file, telemOut, yamlOut)
	if err != nil {
		return &IRacing{}, err
	}

	return &IRacing{
		logger:   logger,
		SDK:      sdk,
		data:     telem.NewTelemetryData(),
		streamCh: make(chan telem.TelemetryData),
		// NOTE: This is because I stupidly recorded a test IBT file in 240
		ticker: time.NewTicker(time.Second / 240),
	}, nil
}

func (i *IRacing) stream(ctx context.Context) {
	i.data.InitialTime = time.Now()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-i.ticker.C:
				i.readData()

				// Publish data
				select {
				case i.streamCh <- *i.data:
				default:
					// skip this data, don't allow publishers to lag behind
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

	// Read binded data
	for _, b := range i.data.ActiveBinds {
		v := i.SDK.Vars.Vars[b.Key].Value

		b.Transform(v, &i.data.Values[b.ID])
	}

	i.data.PenultimateDataPoll = i.data.LastDataPoll
	i.data.LastDataPoll = time.Now()
}

// Telemetry Provider Interface

// Stream returns a channel that we will use to funnel the telemetry data back to the
// UI, which then should broadcast it to the devices
func (i *IRacing) Stream() (<-chan telem.TelemetryData, error) {
	var ctx context.Context
	ctx, i.streamCancel = context.WithCancel(context.Background())

	// Start the stream
	i.stream(ctx)

	return i.streamCh, nil
}

func (i *IRacing) Subscribe(requestFields map[int16]telem.FieldID) {
	i.logger.Debug(fmt.Sprintf("Len Req: %d\n", len(requestFields)))

	i.data.ActiveBinds = make([]telem.BoundField, 0, len(requestFields))

	boundCheck := make(map[telem.FieldID]bool)

	for winID, id := range requestFields {
		i.data.Values[id].IDs = append(i.data.Values[id].IDs, winID)

		// Check if we already bound this FieldID
		if boundCheck[id] {
			continue
		}

		// Translate the UI FieldIDs to this provider's field names
		sdkKey, ok := internalToSDKFieldNames[id]
		if !ok {
			i.logger.Debug("failed to get internal id")
			// Need to find a way to pass a message saying something wasn't right
			continue
		}

		binding := telem.BoundField{
			Key: sdkKey,
			ID:  id,
		}

		switch id {
		case telem.Speed:
			binding.Transform = func(v any, out *telem.TelemetryField) {
				out.Type = telem.DataTypeUINT16
				out.Raw = uint64(conv.MsToKph(v.(float32)))
			}
		case telem.Gear:
			binding.Transform = func(v any, out *telem.TelemetryField) {
				out.Type = telem.DataTypeCHAR
				// out.Raw = uint64(v.(int))

				gear := 0
				if val, ok := v.(int32); ok {
					gear = int(val)
				} else if val, ok := v.(int); ok {
					gear = val
				}

				switch {
				case gear == 0:
					out.Raw = uint64('N') // ASCII 78
				case gear < 0:
					out.Raw = uint64('R') // ASCII 82
				case gear > 0 && gear < 10:
					// Quickest way to turn 1 into '1', 2 into '2', etc.
					// ASCII '0' is 48, so 48 + 1 = 49 ('1')
					out.Raw = uint64('0' + gear)
				default:
					out.Raw = uint64('?') // Fallback
				}
			}
		case telem.RPM:
			binding.Transform = func(v any, out *telem.TelemetryField) {
				out.Type = telem.DataTypeUINT16
				out.Raw = uint64(uint16(v.(float32)))
			}
		case telem.PitSpeedLimiter:
			binding.Transform = PitSpeedLimiterTransform
		case telem.BrakeBias:
			binding.Transform = FloatToStringTransform
		case telem.ABSSetting:
			binding.Transform = FloatToUInt8Transform
		case telem.TCSetting:
			binding.Transform = FloatToUInt8Transform
		case telem.ThrottleSetting:
			binding.Transform = FloatToUInt8Transform
		case telem.LFtempM:
			binding.Transform = func(v any, out *telem.TelemetryField) {
				out.Type = telem.DataTypeSTRING
				out.Str = strconv.FormatFloat(float64(v.(float32)), 'f', 1, 32)
			}
		case telem.SessionTime:
			binding.Transform = func(v any, out *telem.TelemetryField) {
				out.Type = telem.DataTypeSTRING
				out.Str = strconv.FormatFloat(v.(float64), 'f', 1, 32)
			}
		case telem.ReplaySessionTime:
			binding.Transform = func(v any, out *telem.TelemetryField) {
				out.Type = telem.DataTypeSTRING
				out.Str = strconv.FormatFloat(v.(float64), 'f', 1, 32)
			}
		case telem.Empty:
			binding.Transform = EmptyTransform
		case telem.LapLastLapTime:
			binding.Transform = LapTimeTransform
		case telem.LapNumber:
			binding.Transform = UInt8Transform
		}

		i.data.ActiveBinds = append(i.data.ActiveBinds, binding)
		boundCheck[id] = true
	}

	i.logger.Debug(fmt.Sprintf("Subscribed: %+v\n", i.data.ActiveBinds))
}
