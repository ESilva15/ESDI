// Package esdi will have the domain logic for our desktop interface
package esdi

import (
	"esdi/sources/iracing"
	"log"
	"os"
	"time"

	"github.com/ESilva15/goirsdk"
	"github.com/tarm/serial"
)

type ESDI struct {
	SerialConfig *serial.Config
	SerialConn   *serial.Port
	irsdk        *goirsdk.IBT
	data         SimulationData
	dataPacket   DataPacket
	// Source       GameSource
}

func (e *ESDI) Close() {
	err := e.SerialConn.Close()
	if err != nil {
		log.Fatalf("couldn't close serial connection: %v", err)
	}
}

func ESDIInit(port string, baud int) (ESDI, error) {
	sConfig := &serial.Config{
		Name:        port,
		Baud:        baud,
		ReadTimeout: time.Millisecond * 1000,
	}

	// Open the serial port
	sPort, err := serial.OpenPort(sConfig)
	if err != nil {
		return ESDI{}, err
	}

	// return ESDI{sConfig, sPort, nil}, nil
	return ESDI{sConfig, sPort, nil, SimulationData{}, DataPacket{}}, nil
	// return ESDI{nil, nil, nil, SimulationData{}, DataPacket{}}, nil
}

func RunLiveTelemetry(port string, output string, session string) {
	esdi, err := ESDIInit(port, 115200)
	if err != nil {
		log.Fatalf("Failed to get Desktop Interface: %v", err)
	}

	// irsdk, err := iracing.Init(nil, outputFile, sessionFile)
	// if err != nil {
	// 	log.Fatalf("Failed to create iRacing interface: %v", err)
	// }

	irsdk, err := goirsdk.Init(nil, output, session)
	if err != nil {
		log.Fatalf("Failed to create irsdk instance: %v\n", err)
	}

	esdi.irsdk = irsdk

	esdi.telemetry()
}

func RunOfflineTelemetry(port string, input string, output string, session string) {
	esdi, err := ESDIInit(port, 115200)
	if err != nil {
		log.Fatalf("Failed to get Desktop Interface: %v", err)
	}

	file, err := os.Open(input)
	if err != nil {
		log.Fatalf("Failed to open IBT file: %v", err)
	}

	irsdk, err := iracing.Init(file, output, session)
	if err != nil {
		log.Fatalf("Failed to create iRacing interface: %v", err)
	}
	// irsdk, err := goirsdk.Init(file, outFile, sessionFile)
	// if err != nil {
	// 	log.Fatalf("Failed to create irsdk instance: %v\n", err)
	// }

	esdi.irsdk = irsdk.SDK

	esdi.telemetry()
}
