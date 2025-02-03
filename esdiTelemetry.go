package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	// "github.com/tarm/serial"
)

type GameSource interface {
	GetData(string) (interface{}, error)
	UpdateData() error
	GetSessionInfo() (interface{}, error)
}

type DataPacket struct {
	Speed           int32
	Gear            int32
	RPM             int32
	LapCount        int32
	LapDistPct      float32
	LapTime         [16]byte // Current lap time
	LapDelta        [16]byte // Delta to selected reference lap
	BestLapTime     [16]byte // Best lap in session
	LastLapTime     [16]byte // Last lap time
	FuelUsageCurLap float32
	FuelPerLap      float32
	FuelPct         float32
	FuelLiters      float32
	FuelTotal       float32 // This will be calculated and passed in Liters
	Position        int32
	Standings       [5]StandingsLine
}

var (
	paddingStandingsLine = StandingsLine{
		DriverName: createDriverName("---"),
		Lap:        0,
		CarIdx:     0,
		LapPct:     0,
		EstTime:    0,
		TimeBehind: 0,
	}
	iniFuelLvl      float32
	fuelLevels      map[int]float32
	lastMessageTime time.Time
	mu              sync.Mutex
)

func msToKph(v float32) int {
	return int((3600 * v) / 1000)
}

func lapTimeRepresentation(t float32) string {
	if t < 0 {
		t = 0
	}

	wholeSeconds := int64(t)
	lapTime := time.Unix(wholeSeconds, int64((t-float32(wholeSeconds))*1e9))

	return lapTime.Format("04:05.000")
}

func lapTimeDeltaRepresentation(t float32) string {
	sign := '-'
	if t < 0 {
		sign = '+'
		t = -1 * t
	}

	if t >= 1 {
		wholeSeconds := int64(t)
		lapTime := time.Unix(wholeSeconds, int64((t-float32(wholeSeconds))*1e9))
		return fmt.Sprintf("%c%s", sign, lapTime.Format("5.00"))
	}

	return fmt.Sprintf("%c.%02d", sign, int64(t*100))
}

func resetTerminal() {
	// Show cursor and clear screen
	fmt.Print("\033[?25h\033[2J\033[H")
}

func (e *ESDI) setupSignalHandlers() chan string {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	done := make(chan string)

	go func() {
		s := <-sigc
		switch s {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP:
			resetTerminal()
		}
		close(done)
		e.Close()
	}()

	return done
}

// func sendData(e *ESDI, s *serial.Port, done <-chan string, errC chan string) {
// 	ticker := time.NewTicker(time.Millisecond * 25)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-done:
// 			return
// 		case <-ticker.C:
// 			mu.Lock()
//
// 			data := DataToSend{
// 				Speed:           e.data.Speed,
// 				Gear:            e.data.Gear,
// 				RPM:             e.data.RPM,
// 				LapCount:        e.data.LapCount,
// 				LapDistPct:      e.data.LapDistPct,
// 				// FuelPerLap:      e.data.FuelPerLap,
// 				// FuelUsageCurLap: e.data.FuelUsageCurLap,
// 			}
//
// 			copy(data.LapTime[:], e.data.LapTime[:])
// 			copy(data.LapDelta[:], e.data.LapDelta[:])
// 			copy(data.BestLapTime[:], e.data.BestLapTime[:])
// 			copy(data.LastLapTime[:], e.data.LastLapTime[:])
//
// 			var buf bytes.Buffer
// 			err := binary.Write(&buf, binary.LittleEndian, data)
//
// 			_, err = s.Write(buf.Bytes())
// 			if err != nil {
// 				log.Printf("Unable to write data: %v", err)
// 				break
// 			}
//
// 			lastMessageTime = time.Now()
// 			mu.Unlock()
// 		}
// 	}
// }

func printData(e *ESDI, done <-chan string) {
	ticker := time.NewTicker(time.Millisecond * 25)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var buffer strings.Builder
			buffer.WriteString("\033[?25l\033[2J\033[H")

			mu.Lock()
			buffer.WriteString("Car data:\n")
			buffer.WriteString(fmt.Sprintf("Gear: %d, RPM: %d, Speed: %d\n\n",
				e.data.Gear, e.data.RPM, e.data.Speed))

			buffer.WriteString("Fuel data:\n")
			buffer.WriteString(fmt.Sprintf("Fuel: %.2fL/%.2fL %.2f/lap [%.2f%%]\n\n", e.data.FuelLiters,
				e.data.FuelTotal, e.data.FuelPerLap, e.data.FuelPct))

			buffer.WriteString("Lap data:\n")
			buffer.WriteString(fmt.Sprintf("LapTime: %s [%s]\n", e.data.LapTime,
				e.data.LapDelta))
			buffer.WriteString(fmt.Sprintf("Best Lap Time: %s\n", e.data.BestLapTime))
			buffer.WriteString(fmt.Sprintf("Last Lap Time: %s\n", e.data.LastLapTime))
			buffer.WriteString(fmt.Sprintf("Lap: %d [%.2f%%]\n\n", e.data.LapCount,
				e.data.LapDistPct))

			buffer.WriteString("Position data:\n")
			buffer.WriteString(fmt.Sprintf("Pos: %d\n", e.data.Position))

			for p, v := range e.data.Standings {
				s := fmt.Sprintf("[%2d] [%2d] %-16s %3d %10f %s\n",
					p+1, v.CarIdx, string(bytes.Trim(v.DriverName[:], "\x00")), v.Lap, v.LapPct, v.TimeBehindString)
				buffer.WriteString(s)
			}
			mu.Unlock()

			// buffer.WriteString("\n" + message)
			fmt.Print(buffer.String())
		}
	}
}

func (e *ESDI) getVehicleData() {
	curGear := e.irsdk.Vars.Vars["Gear"].Value
	curRPM := e.irsdk.Vars.Vars["RPM"].Value
	curSpeed := e.irsdk.Vars.Vars["Speed"].Value

	mu.Lock()
	e.data.Gear = int32(curGear.(int))
	e.data.RPM = int32(curRPM.(float32))
	e.data.Speed = int32(msToKph(curSpeed.(float32)))
	mu.Unlock()
}

func (e *ESDI) fuelData() {
	fLiters := e.irsdk.Vars.Vars["FuelLevel"].Value
	fPct := e.irsdk.Vars.Vars["FuelLevelPct"].Value

	curLap := e.irsdk.Vars.Vars["Lap"].Value.(int)
	fuelLevels[curLap] = fLiters.(float32)

	if curLap-2 < 0 {
		e.data.FuelPerLap = 0.0
	} else {
		e.data.FuelPerLap = fuelLevels[curLap-2] - fuelLevels[curLap-1]
	}

	mu.Lock()
	e.data.FuelPct = float32(fPct.(float32)) * 100
	e.data.FuelLiters = float32(fLiters.(float32))
	e.data.FuelTotal = (100 * e.data.FuelLiters) / e.data.FuelPct
	mu.Unlock()
}

func (e *ESDI) lapData() {
	currentLap := e.irsdk.Vars.Vars["Lap"].Value
	lapDistPct := e.irsdk.Vars.Vars["LapDistPct"].Value
	currentLapTime := e.irsdk.Vars.Vars["LapCurrentLapTime"].Value
	lapBestLapTime := e.irsdk.Vars.Vars["LapBestLapTime"].Value
	lapLastLapTime := e.irsdk.Vars.Vars["LapLastLapTime"].Value
	lapDeltaToBestLap := e.irsdk.Vars.Vars["LapDeltaToBestLap"].Value

	mu.Lock()
	e.data.LapCount = int32(currentLap.(int))
	e.data.LapDistPct = float32(lapDistPct.(float32)) * 100
	copy(e.data.LapTime[:], string(lapTimeRepresentation(currentLapTime.(float32))))
	copy(e.data.LastLapTime[:], string(lapTimeRepresentation(lapLastLapTime.(float32))))
	copy(e.data.BestLapTime[:], string(lapTimeRepresentation(lapBestLapTime.(float32))))
	copy(e.data.LapDelta[:], string(lapTimeDeltaRepresentation(lapDeltaToBestLap.(float32))))
	mu.Unlock()
}

func (e *ESDI) positionData() {
	standings := createStandingsTable(e.irsdk)
	relativeStandings(e.irsdk, standings, e.irsdk.SessionInfo.DriverInfo.DriverCarIdx)
	p := findEntry(standings, func(l StandingsLine) bool {
		return l.CarIdx == int32(e.irsdk.SessionInfo.DriverInfo.DriverCarIdx)
	})

	lowerLim := p - 2
	upperLim := p + 3

	var lowerPadding []StandingsLine
	var upperPadding []StandingsLine

	if lowerLim < 0 {
		lowerPadding = make([]StandingsLine, abs(lowerLim))
		for k := 0; k < abs(lowerLim); k++ {
			lowerPadding[k] = paddingStandingsLine
		}
		lowerLim = 0
	}
	if upperLim >= len(standings) {
		upperPadding = make([]StandingsLine, upperLim-len(standings))
		for k := 0; k < upperLim-len(standings); k++ {
			upperPadding[k] = paddingStandingsLine
		}
		upperLim = len(standings)
	}

	standings = append(lowerPadding, standings[lowerLim:upperLim]...)
	standings = append(standings, upperPadding...)

	mu.Lock()
	copy(e.data.Standings[:], standings[0:5])
	e.data.Position = int32(p)
	mu.Unlock()
}

type DataToSend struct {
	Speed           int32
	Gear            int32
	RPM             int32
	LapCount        int32
	LapDistPct      float32
	LapTime         [16]byte // Current lap time
	LapDelta        [16]byte // Delta to selected reference lap
	BestLapTime     [16]byte // Best lap in session
	LastLapTime     [16]byte // Last lap time
	FuelUsageCurLap float32
	FuelPerLap      float32
	// FuelPct         float32
	// FuelLiters      float32
	// FuelTotal       float32 // This will be calculated and passed in Liters
	// Position        int32
	// Standings       [5]StandingsLine
}

func (e *ESDI) telemetry() {
	// Set the handlers
	done := e.setupSignalHandlers()

	dataError := make(chan string)

	// go sendData(e, e.SerialConn, done, dataError)
	go printData(e, done)

	mainLoopTicker := time.NewTicker(time.Second / 60)
	defer mainLoopTicker.Stop()

	fuelLevels = make(map[int]float32, 256)

	for {
		select {
		case s := <-done:
			resetTerminal()
			fmt.Println(s)
			return
		case s := <-dataError:
			done <- s
		case <-mainLoopTicker.C:
			time.Sleep(time.Second / 60)

			var err error

			_, err = e.irsdk.Update(time.Millisecond * 100)
			if err != nil {
				fmt.Printf("could not update data: %v", err)
				continue
			}

			e.getVehicleData()
			e.fuelData()
			e.lapData()
			e.positionData()

			mu.Lock()

			data := DataToSend{
				Speed:           e.data.Speed,
				Gear:            e.data.Gear,
				RPM:             e.data.RPM,
				LapCount:        e.data.LapCount,
				LapDistPct:      e.data.LapDistPct,
				// FuelPerLap:      e.data.FuelPerLap,
				// FuelUsageCurLap: e.data.FuelUsageCurLap,
			}

			copy(data.LapTime[:], e.data.LapTime[:])
			copy(data.LapDelta[:], e.data.LapDelta[:])
			copy(data.BestLapTime[:], e.data.BestLapTime[:])
			copy(data.LastLapTime[:], e.data.LastLapTime[:])

			var buf bytes.Buffer
			err = binary.Write(&buf, binary.LittleEndian, data)

      if len(buf.Bytes()) != 92 {
        panic(fmt.Sprintf("Size is: %d, instead of 92", len(buf.Bytes())))
      }

			_, err = e.SerialConn.Write(buf.Bytes())
			if err != nil {
				log.Printf("Unable to write data: %v", err)
				break
			}

			lastMessageTime = time.Now()
			mu.Unlock()
		}
	}
}
