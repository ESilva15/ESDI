package main

import (
	"bytes"
	"encoding/binary"
	"esdi/sources/iracing"
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

func main() {
	// Set the handlers
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
  done := make(chan struct{})

	if len(os.Args) == 3 {
		fmt.Printf("Port: %s\n", os.Args[1])
		fmt.Printf("File: %s\n", os.Args[2])
	} else {
		log.Fatal("Wrong usage.")
	}

	esdi, err := ESDIInit(os.Args[1], 115200)
	if err != nil {
		log.Fatalf("Failed to get Desktop Interface: %v", err)
	}

	file, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatalf("Failed to open IBT file: %v", err)
	}

	irsdk, err := iracing.Init(file)
	if err != nil {
		log.Fatalf("Failed to create iRacing interface: %v", err)
	}

	esdi.Source = &irsdk

	go func() {
		s := <-sigc
		switch s {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP:
      fmt.Printf("Received signal: %v\n", s)
			fmt.Print("\033[?25h\033[2J\033[H")
		}
    close(done)
    esdi.Close()
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

			err = esdi.Source.UpdateData()
			if err != nil {
				fmt.Printf("could not update data: %v", err)
				continue
			}

			curGear, err := esdi.Source.GetData("Gear")
			if err != nil {
				log.Fatalf("could not get field `Gear`: %v", err)
			}

			curRPM, err := esdi.Source.GetData("RPM")
			if err != nil {
				log.Fatalf("could not get field `RPM`: %v", err)
			}

			curSpeed, err := esdi.Source.GetData("Speed")
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

				_, err = esdi.SerialConn.Write(buf.Bytes())
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

	esdi.Close()
}
