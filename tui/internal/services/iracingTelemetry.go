package services

// NOTE: I have to think about how to implement this, for now I'm bruteforcing
// iracing support only. But I should have a generic service that can gather
// telemetry from multiples sims and just publish it in an internal format

import (
	"context"
	esdi "esdi/oldEsdi"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ESilva15/goirsdk"
)

// Car data lengths
const (
	SpeedLen     = 5
	GearLen      = 3
	RpmLen       = 6
	BrakeBiasLen = 6
)

type StreamState int

const (
	StreamStatePaused  StreamState = 0
	StreamStateRunning StreamState = 1
	StreamStateOff     StreamState = 2
)

type IRacingService struct {
	Message         chan string
	LastMessageTime time.Time
	Data            *esdi.SimulationData
	DataView        *esdi.DataPacket
	InitialTime     time.Time
	LastTime        time.Time
	Mut             sync.Mutex
	ticker          *time.Ticker
	Irsdk           *goirsdk.IBT
	//
	isRunning    bool
	Stream       chan string
	StreamCancel context.CancelFunc
}

func NewIRacingService(msg chan string) *IRacingService {
	// Open the telemetry file
	file, err := os.Open("/home/esilva/Desktop/projetos/simracing_peripherals/testTelemetry/supercars_indianapolis.ibt")
	if err != nil {
		log.Fatalf("Failed to open IBT file: %v", err)
	}

	irsdk, err := goirsdk.Init(file, "./out.ibt", "./out.yaml")
	if err != nil {
		log.Fatalf("Failed to load iRacing data")
	}

	return &IRacingService{
		Message:   msg,
		ticker:    time.NewTicker(time.Second / 60),
		Irsdk:     irsdk,
		Stream:    make(chan string, 10),
		Data:      &esdi.SimulationData{},
		DataView:  &esdi.DataPacket{},
		isRunning: false,
	}
}

func (irs *IRacingService) GetStream() <-chan string {
	return irs.Stream
}

func (irs *IRacingService) StartStream() {
	if irs.isRunning {
		return
	}

	var ctx context.Context
	ctx, irs.StreamCancel = context.WithCancel(context.Background())

	irs.startStream(ctx)
	irs.isRunning = true
}

func (irs *IRacingService) StopStream() {
	if irs.StreamCancel != nil {
		irs.StreamCancel()
	}
}

func (irs *IRacingService) startStream(ctx context.Context) {
	irs.Message <- "STARTING THIS\n"
	irs.InitialTime = time.Now()
	go func() {
		for {
			select {
			case <-ctx.Done():
				irs.Message <- "CONTEXT SAID WE ARE DONE\n"
				return
			case <-irs.ticker.C:
				irs.Message <- "GOT A TICK\n"
				irs.ReadData(ctx)
				irs.Stream <- irs.Stringified()
			}
			irs.Message <- "we are we going???\n"
		}
	}()
}

func (irs *IRacingService) ReadData(ctx context.Context) {
	irs.Mut.Lock()
	defer irs.Mut.Unlock()

	var err error

	_, err = irs.Irsdk.Update(time.Millisecond * 100)
	if err != nil {
		return
	}

	irs.getVehicleData()

	// Test the actual dataPacket we are sending over the wire
	copyBytes(irs.DataView.Speed[:], SpeedLen, fmt.Sprintf("%3d", irs.Data.Speed))
	copyBytes(irs.DataView.Gear[:], GearLen, fmt.Sprintf("%2d", irs.Data.Gear))
	copyBytes(irs.DataView.RPM[:], RpmLen, fmt.Sprintf("%3d", irs.Data.RPM))

	irs.LastMessageTime = time.Now()
}

func (irs *IRacingService) getVehicleData() {
	curGear := irs.Irsdk.Vars.Vars["Gear"].Value
	curRPM := irs.Irsdk.Vars.Vars["RPM"].Value
	curSpeed := irs.Irsdk.Vars.Vars["Speed"].Value

	irs.Data.Gear = int32(curGear.(int))
	irs.Data.RPM = int32(curRPM.(float32))
	irs.Data.Speed = int32(msToKph(curSpeed.(float32)))
}

func copyBytes(dest []byte, destSize int, src string) {
	copy(dest[:], []byte(src))
	dest[min(destSize-1, len(src))] = '\x00'
}

func msToKph(v float32) int {
	return int((3600 * v) / 1000)
}

func (irs *IRacingService) Stringified() string {
	var buffer strings.Builder

	irs.Mut.Lock()
	sessionTimeR := irs.Irsdk.Vars.Vars["SessionTime"].Value
	sessionTime := float64(sessionTimeR.(float64))

	currTime := time.Now()
	delta := currTime.Sub(irs.LastTime)
	buffer.WriteString(fmt.Sprintf("[%s]\n", currTime.Format("2006/01/02 15:04:05.000")))
	buffer.WriteString(fmt.Sprintf("Delta: %d [%f]\n\n", delta.Milliseconds(), 1000.0/60.0))
	irs.LastTime = currTime

	elapsed := currTime.Sub(irs.InitialTime)
	softwareElapsed := time.Unix(0, 0).Add(elapsed).Format("04:05.000")
	sessionElapsed := time.Unix(0, 0).
		Add(time.Duration(sessionTime * float64(time.Second))).
		Format("04:05.000")

	buffer.WriteString(fmt.Sprintf("Elapsed (software): %s\n",
		softwareElapsed))
	buffer.WriteString(fmt.Sprintf("Elapsed (session):  %s\n\n",
		sessionElapsed))

	// buffer.WriteString("Car data:\n")
	buffer.WriteString(fmt.Sprintf("Gear: %d, RPM: %d, Speed: %d\n\n",
		irs.Data.Gear, irs.Data.RPM, irs.Data.Speed))

	// buffer.WriteString("Fuel data:\n")
	// buffer.WriteString(fmt.Sprintf("Fuel Est: %s\n\n", e.dataPacket.FuelEst))

	// buffer.WriteString("Lap data:\n")
	// buffer.WriteString(fmt.Sprintf("Delta:         [%s] [%f] [%s]\n", e.dataPacket.DeltaToBestLap,
	// 	e.data.LapDeltaFloat, lapTimeDeltaRepresentation(e.data.LapDeltaFloat)))
	// buffer.WriteString(fmt.Sprintf("LapTime:       %s\n", e.dataPacket.CurrLapTime))
	// buffer.WriteString(fmt.Sprintf("Best Lap Time: %s\n", e.dataPacket.BestLapTime))
	// buffer.WriteString(fmt.Sprintf("Last Lap Time: %s\n", e.dataPacket.LastLapTime))
	// buffer.WriteString(fmt.Sprintf("LapBestNLapTi: %f\n\n", e.data.LapBestNLapTime))

	// buffer.WriteString("Position data:\n")
	// buffer.WriteString(fmt.Sprintf("Pos: %d\n", e.dataPacket.Position))

	// for p, v := range e.dataPacket.Standings {
	// 	s := fmt.Sprintf("[%2d] %s %-16s %-16s\n",
	// 		p+1, v.Lap, string(bytes.Trim(v.DriverName[:], "\x00")), v.TimeBehindString)
	// 	buffer.WriteString(s)
	// }

	// buffer.WriteString(fmt.Sprintf("Size:     %v\n", binary.Size(DataPacket{})))
	// buffer.WriteString(fmt.Sprintf("Recv:     %d\n", e.data.Recv))
	// buffer.WriteString(fmt.Sprintf("Recv Err: %v\n", e.data.ReadError))

	irs.Mut.Unlock()

	return buffer.String()
}
