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

type PeripheralDevice struct {
	Port string
	Name [32]byte
}

type PeripheralDeviceClerk struct {
	Devices map[string]PeripheralDevice
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
		fmt.Println(p)
	}

	return nil
}

// Add a peripheral device manager
// Something that auto discovers available devices
