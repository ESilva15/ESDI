package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type GameSource interface {
	GetData(string) (interface{}, error)
	UpdateData() error
}

type DataPacket struct {
	Speed int32
	Gear  int32
	RPM   int32
}

func msToKph(v float32) int {
	return int((3600 * v) / 1000)
}

func (e *ESDI) telemetry() {
	// Set the handlers
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
			fmt.Printf("Received signal: %v\n", s)
			fmt.Print("\033[?25h\033[2J\033[H")
		}
		close(done)
		e.Close()
	}()

	lastTime := time.Now().UnixMilli()
	lastDataSent := time.Now().UnixMilli()
	for {
		select {
		case <-done:
			return
		default:
			time.Sleep(time.Second / 60)

			var err error
			var buffer strings.Builder
			buffer.WriteString("\033[?25l\033[2J\033[H")

			err = e.Source.UpdateData()
			if err != nil {
				fmt.Printf("could not update data: %v", err)
				continue
			}

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

			buffer.WriteString(fmt.Sprintf("Gear: %d, RPM: %d, Speed: %d", gear, rpm, speed))

			curTime := time.Now().UnixMilli()
			message := fmt.Sprintf("%d,%d\n", gear-1, rpm)
			buffer.WriteString("\n" + message)

			messageWasSentMark := "N"
			if curTime-lastDataSent > 25 {
				packet := DataPacket{
					Speed: speed,
					Gear:  gear,
					RPM:   rpm,
				}

				var buf bytes.Buffer
				err = binary.Write(&buf, binary.LittleEndian, packet)

				_, err = e.SerialConn.Write(buf.Bytes())
				if err != nil {
					log.Printf("Unable to write data: %v", err)
					break
				}

				messageWasSentMark = "Y"
				lastDataSent = curTime
			}

			buffer.WriteString(" -> " + messageWasSentMark)
			if curTime-lastTime > 100 {
				fmt.Print(buffer.String())
				lastTime = curTime
			}

		}
	}

	e.Close()
}
