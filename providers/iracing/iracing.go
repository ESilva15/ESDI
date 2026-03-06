package iracing

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	conv "esdi/conversions"
	telem "esdi/telemetry"

	"github.com/ESilva15/goirsdk"
)

// IRacing is our iRacing telemetry data provider - its a TelemetryProvider interface
type IRacing struct {
	SDK *goirsdk.IBT

	// Data Handling
	mut  sync.Mutex
	data *telem.TelemetryData

	// Timing information
	ticker *time.Ticker // ticker will keep polling intervals constant

	// Stream
	streamCh     chan telem.TelemetryData
	streamCancel context.CancelFunc
	// isRunning bool
}

func NewIRacingProvider(
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
				i.ReadData()

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

func (i *IRacing) ReadData() {
	i.mut.Lock()
	defer i.mut.Unlock()

	var err error

	_, err = i.SDK.Update(time.Millisecond * 100)
	if err != nil {
		return
	}

	i.readVehicleData()

	i.data.PenultimateDataPoll = i.data.LastDataPoll
	i.data.LastDataPoll = time.Now()
}

func (i *IRacing) readVehicleData() {
	curGear := i.SDK.Vars.Vars["Gear"].Value
	curRPM := i.SDK.Vars.Vars["RPM"].Value
	curSpeed := i.SDK.Vars.Vars["Speed"].Value

	speed := conv.MsToKph(curSpeed.(float32))
	gear := uint8(curGear.(int))
	rpm := uint16(curRPM.(float32))

	i.data.Values[telem.Speed] = telem.TelemetryField{
		Type: telem.DataTypeUINT16,
		Raw:  uint64(speed),
	}

	i.data.Values[telem.Gear] = telem.TelemetryField{
		Type: telem.DataTypeUINT8,
		Raw:  uint64(gear),
	}

	i.data.Values[telem.RPM] = telem.TelemetryField{
		Type: telem.DataTypeUINT16,
		Raw:  uint64(rpm),
	}
}

func (i *IRacing) Stream() (<-chan telem.TelemetryData, error) {
	var ctx context.Context
	ctx, i.streamCancel = context.WithCancel(context.Background())

	// Start the stream
	i.stream(ctx)

	return i.streamCh, nil
}
