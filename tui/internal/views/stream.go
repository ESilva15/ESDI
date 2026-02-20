package views

import (
	esdi "esdi/oldEsdi"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ESilva15/goirsdk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Car data lengths
const (
	SpeedLen     = 5
	GearLen      = 3
	RpmLen       = 6
	BrakeBiasLen = 6
)

const (
	streamingBoxID = "streaming-box"
)

var (
	lastMessageTime time.Time
	data            *esdi.SimulationData = &esdi.SimulationData{}
	dataView        *esdi.DataPacket     = &esdi.DataPacket{}
	initialTime                          = time.Now()
	lastTime                             = time.Now()
	mu              sync.Mutex
	irsdk           *goirsdk.IBT
)

func setUpSource(bus *events.Bus, doc *dom.DOM, tv *tview.TextView) {
	// Open the telemetry file
	file, err := os.Open("/home/esilva/Desktop/projetos/simracing_peripherals/testTelemetry/supercars_indianapolis.ibt")
	if err != nil {
		log.Fatalf("Failed to open IBT file: %v", err)
	}

	irsdk, err = goirsdk.Init(file, "./out.ibt", "./out.yaml")
	if err != nil {
		log.Fatalf("Failed to load iRacing data")
	}

	startTime := time.Now()

	go readData()

	// global irsdk is now here
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		counter := 0
		for t := range ticker.C {
			// whatever
			currentTime := time.Now()

			delta := currentTime.Sub(startTime)

			str := stringified() + fmt.Sprintf("%v - %7d\n%5f - %5f",
				t.UTC(), counter, delta.Seconds(), delta.Seconds()/float64(counter))

			tv.SetText(str)
			bus.Emit(ui.ForceRedraw{})
			counter++
		}
	}()
}

func streamingWindow(bus *events.Bus, doc *dom.DOM) {
	bus.Emit(ui.LogEv{Log: "creating streaming window"})

	// Get the ActionPages parent
	box := tview.NewTextView()
	box.SetBorder(true).SetTitle("STREAMING")
	box.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Rune() {
		case 'q':
			// well, I really need to design the UI to work properly
			os.Exit(0)
		}

		return ev
	})

	var err error
	var boxNode *dom.UINode

	// delete the currently existing streaming-box
	actionPages := doc.GetElemByID(layoutToolActionPagesID).(*tview.Pages)
	actionPages.RemovePage(streamingBoxID)

	// Get the currently exisiting streaming box in the dom so we can delete it
	boxNode = doc.GetNodeByID(streamingBoxID)
	if boxNode != nil {
		doc.DeleteElem(boxNode)
	}

	// Register this streaming box as an UINode
	boxNode, err = doc.NewUINode(
		streamingBoxID,
		doc.GetElemByID(layoutToolActionPagesID),
		box,
	)
	if err != nil {
		return
	}

	// Add it to the action pages and view it
	AddAndShowPage(bus, doc, actionPages, boxNode, true)

	setUpSource(bus, doc, box)
}

func stringified() string {
	var buffer strings.Builder

	mu.Lock()
	sessionTimeR := irsdk.Vars.Vars["SessionTime"].Value
	sessionTime := float64(sessionTimeR.(float64))

	currTime := time.Now()
	delta := currTime.Sub(lastTime)
	buffer.WriteString(fmt.Sprintf("[%s]\n", currTime.Format("2006/01/02 15:04:05.000")))
	buffer.WriteString(fmt.Sprintf("Delta: %d [%f]\n\n", delta.Milliseconds(), 1000.0/60.0))
	lastTime = currTime

	elapsed := currTime.Sub(initialTime)
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
		data.Gear, data.RPM, data.Speed))

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

	mu.Unlock()

	return buffer.String()
}

func msToKph(v float32) int {
	return int((3600 * v) / 1000)
}

func lapTimeRepresentation(t float32, f string) string {
	if t < 0 {
		t = 0
	}

	wholeSeconds := int64(t)
	lapTime := time.Unix(wholeSeconds, int64((t-float32(wholeSeconds))*1e9))

	return lapTime.Format(f)
}

func lapTimeDeltaRepresentation(t float32) string {
	sign := '-'
	if t < 0 {
		sign = '+'
		t = -t
	}

	// Cap to 99.9 max
	if t > 99.9 {
		t = 99.9
	}

	if t >= 1 {
		// Round to nearest tenth
		rounded := float32(int(t*10+0.5)) / 10
		return fmt.Sprintf("%c%.1f", sign, rounded)
	}

	// t < 1: round to nearest tenth and remove leading zero (e.g., "0.1" → ".1")
	rounded := float32(int(t*10+0.5)) / 10
	s := fmt.Sprintf("%.1f", rounded)
	if strings.HasPrefix(s, "0") {
		s = s[1:]
	}
	return fmt.Sprintf("%c%s", sign, s)
}

func readData() {
	dataReaderTicker := time.NewTicker(time.Second / 60)
	defer dataReaderTicker.Stop()

	initialTime = time.Now()
	for _ = range dataReaderTicker.C {
		var err error

		mu.Lock()
		_, err = irsdk.Update(time.Millisecond * 100)
		if err != nil {
			fmt.Printf("could not update data: %v", err)
			continue
		}
		mu.Unlock()

		getVehicleData()

		// Test the actual dataPacket we are sending over the wire
		copyBytes(dataView.Speed[:], SpeedLen, fmt.Sprintf("%3d", data.Speed))
		copyBytes(dataView.Gear[:], GearLen, fmt.Sprintf("%2d", data.Gear))
		copyBytes(dataView.RPM[:], RpmLen, fmt.Sprintf("%3d", data.RPM))

		lastMessageTime = time.Now()
	}
}

func getVehicleData() {
	mu.Lock()
	curGear := irsdk.Vars.Vars["Gear"].Value
	curRPM := irsdk.Vars.Vars["RPM"].Value
	curSpeed := irsdk.Vars.Vars["Speed"].Value

	data.Gear = int32(curGear.(int))
	data.RPM = int32(curRPM.(float32))
	data.Speed = int32(msToKph(curSpeed.(float32)))
	mu.Unlock()
}

func copyBytes(dest []byte, destSize int, src string) {
	copy(dest[:], []byte(src))
	dest[min(destSize-1, len(src))] = '\x00'
}
