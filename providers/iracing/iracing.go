package iracing

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	conv "esdi/conversions"
	helper "esdi/helpers"
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
	initialTime     time.Time
	lastMessageTime time.Time
	ticker          time.Ticker // ticker will keep polling intervals constant

	// Stream
	streamCh     chan telem.TelemetryData
	streamCancel context.CancelFunc
	// isRunning bool
}

func NewIRacingProvider(source string, telemOut string, yamlOut string) (*IRacing, error) {
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
		ticker:   *time.NewTicker(time.Second / 60),
	}, nil
}

func (i *IRacing) stream(ctx context.Context) {
	i.initialTime = time.Now()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-i.ticker.C:
				i.ReadData()

				// Publish data
				i.streamCh <- *i.data
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

	i.lastMessageTime = time.Now()
}

func (i *IRacing) readVehicleData() {
	curGear := i.SDK.Vars.Vars["Gear"].Value
	curRPM := i.SDK.Vars.Vars["RPM"].Value
	curSpeed := i.SDK.Vars.Vars["Speed"].Value

	speed := fmt.Sprintf("%3d", int32(conv.MsToKph(curSpeed.(float32))))
	gear := fmt.Sprintf("%2d", int32(curGear.(int)))
	rpm := fmt.Sprintf("%3d", int32(curRPM.(float32)))

	speedArray := i.data.Values["Speed"]
	helper.CopyBytes(speedArray[:], speed)
	i.data.Values["Speed"] = speedArray

	gearArray := i.data.Values["Gear"]
	helper.CopyBytes(gearArray[:], gear)
	i.data.Values["Gear"] = gearArray

	rpmArray := i.data.Values["RPM"]
	helper.CopyBytes(rpmArray[:], rpm)
	i.data.Values["RPM"] = rpmArray
}

func (i *IRacing) Stream() (<-chan telem.TelemetryData, error) {
	var ctx context.Context
	ctx, i.streamCancel = context.WithCancel(context.Background())

	// Start the stream
	i.stream(ctx)

	return i.streamCh, nil
}
