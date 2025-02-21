package main

import (
	"log"

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
		Name: port,
		Baud: baud,
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
