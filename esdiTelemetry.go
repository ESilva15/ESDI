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
	LapTime     string // Current lap time
	LapDelta    string // Delta to selected reference lap
	BestLapTime string // Best lap in session
	LastLapTime string // Last lap time
	FuelPct     float32
	FuelLiters  float32
	FuelTotal   float32 // This will be calculated and passed in Liters
	Position    int32
}

var (
	data            DataPacket
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

func sendData(s *serial.Port, done <-chan struct{}) {
	ticker := time.NewTicker(time.Millisecond * 25)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			mu.Lock()

			var buf bytes.Buffer
			err := binary.Write(&buf, binary.LittleEndian, data)

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

func printData(done <-chan struct{}) {
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
			// message := fmt.Sprintf("%d,%d\n", data.Gear-1, data.RPM)
			buffer.WriteString(fmt.Sprintf("Gear: %d, RPM: %d, Speed: %d\n",
				data.Gear, data.RPM, data.Speed))
			buffer.WriteString(fmt.Sprintf("Fuel: %.2fL/%.2fL [%.2f%%]\n", data.FuelLiters,
				data.FuelTotal, data.FuelPct))
			buffer.WriteString(fmt.Sprintf("LapTime: %s [%s]\n", data.LapTime,
				data.LapDelta))
			buffer.WriteString(fmt.Sprintf("Best Lap Time: %s\n", data.BestLapTime))
			buffer.WriteString(fmt.Sprintf("Last Lap Time: %s\n", data.LastLapTime))
			buffer.WriteString(fmt.Sprintf("Lap: %d [%.2f%%]\n", data.LapCount,
				data.LapDistPct))
			buffer.WriteString(fmt.Sprintf("Pos: %d\n", data.Position))
			mu.Unlock()

			// buffer.WriteString("\n" + message)
			fmt.Print(buffer.String())
		}
	}
}

func (e *ESDI) telemetry() {
	// Set the handlers
	done := e.setupSignalHandlers()

	// go sendData(e.SerialConn, done)
	go printData(done)

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

			err = e.Source.UpdateData()
			if err != nil {
				fmt.Printf("could not update data: %v", err)
				continue
			}

			// Vehicle Movement data
			curGear, err := e.Source.GetData("Gear")
			if err != nil {
				log.Fatalf("could not get field `Gear`: %v", err)
			}

			curRPM, err := e.Source.GetData("RPM")
			if err != nil {
				log.Fatalf("could not get field `RPM`: %v", err)
			}

			curSpeed, err := e.Source.GetData("Speed")
			if err != nil {
				log.Fatalf("could not get field `Speed`: %v", err)
			}

			gear := int32(curGear.(int))
			rpm := int32(curRPM.(float32))
			speed := int32(msToKph(curSpeed.(float32)))

			// Fuel data
			fLiters, err := e.Source.GetData("FuelLevel")
			if err != nil {
				log.Fatalf("could not get field `FuelLevel`: %v", err)
			}

			fPct, err := e.Source.GetData("FuelLevelPct")
			if err != nil {
				log.Fatalf("could not get field `FuelLevelPct`: %v", err)
			}

			fuelPct := float32(fPct.(float32)) * 100
			fuelLiters := float32(fLiters.(float32))
			totalFuel := (100 * fuelLiters) / fuelPct

			// Lap data
			currentLap, err := e.Source.GetData("Lap")
			if err != nil {
				log.Fatalf("could not get field `Lap`: %v", err)
			}

			lapDistPct, err := e.Source.GetData("LapDistPct")
			if err != nil {
				log.Fatalf("could not get field `Lap`: %v", err)
			}

			currentLapTime, err := e.Source.GetData("LapCurrentLapTime")
			if err != nil {
				log.Fatalf("could not get field `LapCurrentLapTime`: %v", err)
			}

			lapBestLapTime, err := e.Source.GetData("LapBestLapTime")
			if err != nil {
				log.Fatalf("could not get field `LapBestLapTime`: %v", err)
			}

			lapLastLapTime, err := e.Source.GetData("LapLastLapTime")
			if err != nil {
				log.Fatalf("could not get field `LapBestLapTime`: %v", err)
			}

			lapDeltaToBestLap, err := e.Source.GetData("LapDeltaToBestLap")
			if err != nil {
				log.Fatalf("could not get field `LapDeltaToBestLap`: %v", err)
			}

			lap := int32(currentLap.(int))
			lapPct := float32(lapDistPct.(float32)) * 100
			lapTime := string(lapTimeRepresentation(currentLapTime.(float32)))
			bestLapTime := string(lapTimeRepresentation(lapBestLapTime.(float32)))
			lapDelta := string(lapTimeDeltaRepresentation(lapDeltaToBestLap.(float32)))
			lastLapTime := string(lapTimeRepresentation(lapLastLapTime.(float32)))

			// Relative data
			// sessionInfoRaw, _ := e.Source.GetSessionInfo()
			// sessionInfo := sessionInfoRaw.(goirsdk.SessionInfoYAML)
			// drivers := sessionInfo.DriverInfo.Drivers
			//
			// selfID := sessionInfo.DriverInfo.DriverCarIdx

			PlayerCarPosition, err := e.Source.GetData("CarIdxPosition")
			if err != nil {
				log.Fatalf("could not get field `CarIdxPosition`: %v", err)
			}

			selfPos := int32(PlayerCarPosition.(int))

			mu.Lock()
			data = DataPacket{
				Speed:       speed,
				Gear:        gear,
				RPM:         rpm,
				LapCount:    lap,
				LapDistPct:  lapPct,
				LapTime:     lapTime,
				LastLapTime: lastLapTime,
				BestLapTime: bestLapTime,
				LapDelta:    lapDelta,
				FuelLiters:  fuelLiters,
				FuelPct:     fuelPct,
				FuelTotal:   totalFuel,
				Position:    selfPos,
			}
			mu.Unlock()
		}
	}
}
