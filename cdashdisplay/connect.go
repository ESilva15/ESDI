package cdashdisplay

import (
	"esdi/peripheral/communication"
	"esdi/peripheral/communication/packets"
	"fmt"
	"path/filepath"
	"time"

	"github.com/tarm/serial"
)

func listPorts() ([]string, error) {
	ttyUSBs, err := filepath.Glob("/dev/ttyUSB*")
	if err != nil {
		return []string{}, err
	}

	return ttyUSBs, nil
}

func probe(WT *communication.WalkieTalkie) error {
	err := WT.TurnOn()
	if err != nil {
		return err
	}

	// Send the identification command
	cmd := communication.CmdRequestID
	var response packets.IdentificationPacket
	err = WT.SendCommand(cmd, []byte{0x06, 0x07, 0x08, 0x09}, &response)
	if err != nil {
		return err
	}

	if response.DeviceID != 0x01 {
		return fmt.Errorf("wrong ID")
	}

	return nil
}

func findDisplayPort() (*communication.WalkieTalkie, error) {
	ports, err := listPorts()
	if err != nil {
		return nil, err
	}

	pLogger.Info(fmt.Sprintf("Looking into %v", ports))

	var wt *communication.WalkieTalkie
	for _, port := range ports {
		wt = &communication.WalkieTalkie{
			Cfg: &serial.Config{
				Name:        port,
				Baud:        115200,
				ReadTimeout: 500 * time.Millisecond,
			},
		}

		err = probe(wt)
		if err == nil {
			break
		}

		pLogger.Info(fmt.Sprintf("wasn't port %s", port))
		wt = nil
	}

	if wt == nil {
		return nil, fmt.Errorf("couldn't find cdashdisplay")
	}

	pLogger.Info(fmt.Sprintf("found cdashdisplay on port: %s", wt.Cfg.Name))
	return wt, nil
}
