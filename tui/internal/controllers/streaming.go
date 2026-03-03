package controllers

import (
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

type StreamingCtrl struct {
	LastMessageTime time.Time
	Data            *esdi.SimulationData
	DataView        *esdi.DataPacket
	InitialTime     time.Time
	LastTime        time.Time
	Mut             sync.Mutex
	Irsdk           *goirsdk.IBT
}

func NewStreamingCtrl() *StreamingCtrl {
	return &StreamingCtrl{
		Data:        &esdi.SimulationData{},
		DataView:    &esdi.DataPacket{},
		InitialTime: time.Now(),
		LastTime:    time.Now(),
	}
}

// func (sc *StreamingCtrl) Stop(bus *events.Bus) {
// }

func (sc *StreamingCtrl) Start() {
	// Open the telemetry file
	file, err := os.Open("/home/esilva/Desktop/projetos/simracing_peripherals/testTelemetry/supercars_indianapolis.ibt")
	if err != nil {
		log.Fatalf("Failed to open IBT file: %v", err)
	}

	sc.Irsdk, err = goirsdk.Init(file, "./out.ibt", "./out.yaml")
	if err != nil {
		log.Fatalf("Failed to load iRacing data")
	}

	startTime := time.Now()

	go sc.ReadData()

	// global irsdk is now here
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		counter := 0
		for t := range ticker.C {
			// whatever
			currentTime := time.Now()

			delta := currentTime.Sub(startTime)

			_ = sc.Stringified() + fmt.Sprintf("%v - %7d\n%5f - %5f",
				t.UTC(), counter, delta.Seconds(), delta.Seconds()/float64(counter))

			// bus.Emit(ui.StreamDataEv{Str: str})
			counter++
		}
	}()
}

func (sc *StreamingCtrl) ReadData() {
	dataReaderTicker := time.NewTicker(time.Second / 60)
	defer dataReaderTicker.Stop()

	sc.InitialTime = time.Now()
	for _ = range dataReaderTicker.C {
		var err error

		sc.Mut.Lock()
		_, err = sc.Irsdk.Update(time.Millisecond * 100)
		if err != nil {
			fmt.Printf("could not update data: %v", err)
			continue
		}
		sc.Mut.Unlock()

		sc.getVehicleData()

		// Test the actual dataPacket we are sending over the wire
		copyBytes(sc.DataView.Speed[:], SpeedLen, fmt.Sprintf("%3d", sc.Data.Speed))
		copyBytes(sc.DataView.Gear[:], GearLen, fmt.Sprintf("%2d", sc.Data.Gear))
		copyBytes(sc.DataView.RPM[:], RpmLen, fmt.Sprintf("%3d", sc.Data.RPM))

		sc.LastMessageTime = time.Now()
	}
}

func (sc *StreamingCtrl) getVehicleData() {
	sc.Mut.Lock()
	curGear := sc.Irsdk.Vars.Vars["Gear"].Value
	curRPM := sc.Irsdk.Vars.Vars["RPM"].Value
	curSpeed := sc.Irsdk.Vars.Vars["Speed"].Value

	sc.Data.Gear = int32(curGear.(int))
	sc.Data.RPM = int32(curRPM.(float32))
	sc.Data.Speed = int32(msToKph(curSpeed.(float32)))
	sc.Mut.Unlock()
}

func copyBytes(dest []byte, destSize int, src string) {
	copy(dest[:], []byte(src))
	dest[min(destSize-1, len(src))] = '\x00'
}

func msToKph(v float32) int {
	return int((3600 * v) / 1000)
}

// Move this function to the views
func (sc *StreamingCtrl) Stringified() string {
	var buffer strings.Builder

	sc.Mut.Lock()
	sessionTimeR := sc.Irsdk.Vars.Vars["SessionTime"].Value
	sessionTime := float64(sessionTimeR.(float64))

	currTime := time.Now()
	delta := currTime.Sub(sc.LastTime)
	buffer.WriteString(fmt.Sprintf("[%s]\n", currTime.Format("2006/01/02 15:04:05.000")))
	buffer.WriteString(fmt.Sprintf("Delta: %d [%f]\n\n", delta.Milliseconds(), 1000.0/60.0))
	sc.LastTime = currTime

	elapsed := currTime.Sub(sc.InitialTime)
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
		sc.Data.Gear, sc.Data.RPM, sc.Data.Speed))

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

	sc.Mut.Unlock()

	return buffer.String()
}
