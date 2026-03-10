package iracing

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
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
	mut            sync.Mutex
	data           *telem.TelemetryData
	activeBindings []boundField

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
		ticker:   time.NewTicker(time.Second / 60),
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

	_, err = i.SDK.Update(time.Millisecond * 100)
	if err != nil {
		return
	}

	// Read binded data
	for _, b := range i.activeBindings {
		v := i.SDK.Vars.Vars[b.Key].Value
		// NOTE: for the love of god, find a way of avoiding this shit
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

func (i *IRacing) Subscribe(requestFields []telem.FieldID) {
	i.logger.Debug(fmt.Sprintf("Len Req: %d\n", len(requestFields)))

	i.activeBindings = make([]boundField, 0, len(requestFields))

	for _, id := range requestFields {
		// Translate the UI FieldIDs to this provider's field names
		sdkKey, ok := internalToSDKFieldNames[id]
		if !ok {
			i.logger.Debug("failed to get internal id")
			// Need to find a way to pass a message saying something wasn't right
			continue
		}

		binding := boundField{
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
				out.Type = telem.DataTypeUINT8
				out.Raw = uint64(v.(int))
			}
		case telem.RPM:
			binding.Transform = func(v any, out *telem.TelemetryField) {
				out.Type = telem.DataTypeUINT16
				out.Raw = uint64(uint16(v.(float32)))
			}
		}

		i.activeBindings = append(i.activeBindings, binding)
	}

	i.logger.Debug(fmt.Sprintf("Subscribed: %+v\n", i.activeBindings))
}
