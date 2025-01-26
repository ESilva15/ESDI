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

	"github.com/tarm/serial"
)

type GameSource interface {
	GetData(string) (interface{}, error)
	UpdateData() error
	GetSessionInfo() (interface{}, error)
}

type DataPacket struct {
	Speed       int32
	Gear        int32
	RPM         int32
	LapCount    int32
	LapDistPct  float32
	LapTime     [16]byte // Current lap time
	LapDelta    [16]byte // Delta to selected reference lap
	BestLapTime [16]byte // Best lap in session
	LastLapTime [16]byte // Last lap time
	FuelPct     float32
	FuelLiters  float32
	FuelTotal   float32 // This will be calculated and passed in Liters
	Position    int32
	Standings   []StandingsLine
}

var (
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

func (e *ESDI) setupSignalHandlers() chan struct{} {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	done := make(chan struct{})

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

func sendData(e *ESDI, s *serial.Port, done <-chan struct{}) {
	ticker := time.NewTicker(time.Millisecond * 25)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			mu.Lock()

			var buf bytes.Buffer
			err := binary.Write(&buf, binary.LittleEndian, e.data)

			_, err = s.Write(buf.Bytes())
			if err != nil {
				log.Printf("Unable to write data: %v", err)
				break
			}

			lastMessageTime = time.Now()
			mu.Unlock()
		}
	}
}

func printData(e *ESDI, done <-chan struct{}) {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var buffer strings.Builder
			buffer.WriteString("\033[?25l\033[2J\033[H")

			mu.Lock()
			// message := fmt.Sprintf("%d,%d\n", e.data.Gear-1, e.data.RPM)
			buffer.WriteString(fmt.Sprintf("Gear: %d, RPM: %d, Speed: %d\n",
				e.data.Gear, e.data.RPM, e.data.Speed))
			buffer.WriteString(fmt.Sprintf("Fuel: %.2fL/%.2fL [%.2f%%]\n", e.data.FuelLiters,
				e.data.FuelTotal, e.data.FuelPct))
			buffer.WriteString(fmt.Sprintf("LapTime: %s [%s]\n", e.data.LapTime,
				e.data.LapDelta))
			buffer.WriteString(fmt.Sprintf("Best Lap Time: %s\n", e.data.BestLapTime))
			buffer.WriteString(fmt.Sprintf("Last Lap Time: %s\n", e.data.LastLapTime))
			buffer.WriteString(fmt.Sprintf("Lap: %d [%.2f%%]\n", e.data.LapCount,
				e.data.LapDistPct))
			buffer.WriteString(fmt.Sprintf("Pos: %d\n", e.data.Position))

			for p, v := range e.data.Standings {
				s := fmt.Sprintf("[%2d] [%2d]%-30s %3d %10f %s\n",
					p+1, v.CarIdx, v.DriverName, v.Lap, v.LapPct, lapTimeRepresentation(v.TimeBehind))
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
		return l.CarIdx == e.irsdk.SessionInfo.DriverInfo.DriverCarIdx
	})

	lowerLim := p - 2
	upperLim := p + 3

	if lowerLim < 0 {
		lowerLim = 0
	}
	if upperLim >= len(standings) {
		upperLim = len(standings)
	}

	standings = standings[lowerLim:upperLim]

	mu.Lock()
	e.data.Standings = standings
	e.data.Position = int32(p)
	mu.Unlock()
}

func (e *ESDI) telemetry() {
	// Set the handlers
	done := e.setupSignalHandlers()

	// go sendData(e.SerialConn, done)
	go printData(e, done)

	mainLoopTicker := time.NewTicker(time.Second / 60)
	defer mainLoopTicker.Stop()

	for {
		select {
		case <-done:
			resetTerminal()
			fmt.Println("Quitting ESDI...")
			return
		case <-mainLoopTicker.C:
			time.Sleep(time.Second / 60)

			var err error

			_, err = e.irsdk.Update(time.Millisecond * 100)
			if err != nil {
				fmt.Printf("could not update data: %v", err)
				continue
			}

			e.getVehicleData()
			e.lapData()
			e.fuelData()
			e.positionData()
		}
	}
}
