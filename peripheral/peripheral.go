// Package peripheral will handle communicating with peripherals
package peripheral

import (
	"fmt"
	"path/filepath"
)

type PeripheralType string

const (
	DisplayPeripheral PeripheralType = "display"
)

type PeripheralDeviceClerk struct {
	// Devices map[string]PeripheralDevice
}

func NewPeripheralDeviceClerk() *PeripheralDeviceClerk {
	return &PeripheralDeviceClerk{}
}

func (clerk *PeripheralDeviceClerk) listPorts() ([]string, error) {
	ttyUSBs, err := filepath.Glob("/dev/ttyUSB*")
	if err != nil {
		return []string{}, err
	}

	return ttyUSBs, nil
}

func (clerk *PeripheralDeviceClerk) FindDevices() error {
	ports, err := clerk.listPorts()
	if err != nil {
		return err
	}

	for _, p := range ports {
		newDevice := NewPeripheralDevice(p)
		err := newDevice.Probe()

		var msg string
		if err == nil {
			msg = string(newDevice.Name[:])
		} else {
			msg = err.Error()
		}

		fmt.Println(p, msg)
	}

	return nil
}
