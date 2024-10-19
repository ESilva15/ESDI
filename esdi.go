package main

import (
	"log"

	"github.com/tarm/serial"
)

type ESDI struct {
	SerialConfig *serial.Config
	SerialConn   *serial.Port
	Source       GameSource
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

	return ESDI{sConfig, sPort, nil}, nil
}
