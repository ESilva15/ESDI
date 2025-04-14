package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	// "encoding/binary"
	"fmt"
	// "log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	// "github.com/tarm/serial"
)

var (
	initialTime          = time.Now()
	lastTime             = time.Now()
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

const (
	LapTimeFormatStr       = "04:05.000"
	RelativeDeltaFormatStr = "04:05.0"
)

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
			fmt.Print("Closing ESDI")
		}
		close(done)
		e.Close()
	}()

	return done
}

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
			sessionTimeR := e.irsdk.Vars.Vars["SessionTime"].Value
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
				e.dataPacket.Gear, e.dataPacket.RPM, e.dataPacket.Speed))

			buffer.WriteString("Fuel data:\n")
			buffer.WriteString(fmt.Sprintf("Fuel Est: %s\n", e.dataPacket.FuelEst))

			buffer.WriteString("Lap data:\n")
			buffer.WriteString(fmt.Sprintf("Delta:         [%s] [%f] [%s]\n", e.dataPacket.DeltaToBestLap,
				e.data.LapDeltaFloat, lapTimeDeltaRepresentation(e.data.LapDeltaFloat)))
			buffer.WriteString(fmt.Sprintf("LapTime:       %s\n", e.dataPacket.CurrLapTime))
			buffer.WriteString(fmt.Sprintf("Best Lap Time: %s\n", e.dataPacket.BestLapTime))
			buffer.WriteString(fmt.Sprintf("Last Lap Time: %s\n", e.dataPacket.LastLapTime))

			// buffer.WriteString("Position data:\n")
			// buffer.WriteString(fmt.Sprintf("Pos: %d\n", e.dataPacket.Position))

			for p, v := range e.dataPacket.Standings {
				s := fmt.Sprintf("[%2d] %s %-16s %-16s\n",
					p+1, v.Lap, string(bytes.Trim(v.DriverName[:], "\x00")), v.TimeBehindString)
				buffer.WriteString(s)
			}

			buffer.WriteString(fmt.Sprintf("Size:     %v\n", binary.Size(DataPacket{})))
			buffer.WriteString(fmt.Sprintf("Recv:     %d\n", e.data.Recv))
			buffer.WriteString(fmt.Sprintf("Recv Err: %v\n", e.data.ReadError))

			mu.Unlock()

			// buffer.WriteString("\n" + message)
			fmt.Print(buffer.String())
		}
	}
}

type DataReq struct {
	Req int8
}

func (e *ESDI) telemetry() {
	// Set the handlers
	done := e.setupSignalHandlers()
	dataError := make(chan string)

	// Display the data on the terminal periodically
	go printData(e, done)

	// Maybe create a struct made to calculate the fuel levels
	// -> struct FuelLvlCalculator
	fuelLevels = make(map[int]float32, 256)

	// We need to add another goroutine here that continuously updates
	// the data
	go readData(e, done)

	// Add another goroutine that is actually waiting for data requests
	// And gives the signal or sends the data
	// go waitForReq(e, done)

	// TODO:
	// Add start and end markers to the data frames

	// This loop will wait for the display to request the data
	total := time.Microsecond * 0
	previousRequest := time.Now()
	nRequests := 0
	for {
		isRunning := true
		select {
		case s := <-done:
			resetTerminal()
			fmt.Println(s)
			isRunning = false
			break
		case s := <-dataError:
			done <- s
		default:
			// Wait for the display to request some data
			var r DataReq
			err := binary.Read(e.SerialConn, binary.LittleEndian, &r)
			e.data.ReadError = err
			if err != nil && err != io.EOF {
				log.Println(err)
				fmt.Println("->", r)
			}
			e.data.Recv = r.Req

			if r.Req == 5 {
				e.SerialConn.Flush()
				currTime := time.Now()
				if nRequests != 0 {
					total += currTime.Sub(previousRequest)
				}
				previousRequest = currTime
				nRequests += 1

				var buf bytes.Buffer
				err = binary.Write(&buf, binary.LittleEndian, e.dataPacket)

				_, err = e.SerialConn.Write(buf.Bytes())
				if err != nil {
					log.Printf("Unable to write data: %v", err)
					break
				}
				e.SerialConn.Flush()
			}
		}

		if !isRunning {
			break
		}
	}

	log.Println("Statistics:")
	log.Printf("  Reqs:             %d\n", nRequests)
	log.Printf("  Total time:       %d (ms)\n", total.Milliseconds())
	log.Printf("  Average req time: %d (ms / request)\n", total.Milliseconds()/int64(nRequests))
}
