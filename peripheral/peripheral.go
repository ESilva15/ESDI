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
	// mu      sync.RWMutex
	Devices map[string]*PeripheralDevice
}

func NewPeripheralDeviceClerk() *PeripheralDeviceClerk {
	return &PeripheralDeviceClerk{
		Devices: make(map[string]*PeripheralDevice),
	}
}

func (clerk *PeripheralDeviceClerk) listPorts() ([]string, error) {
	ttyUSBs, err := filepath.Glob("/dev/ttyUSB*")
	if err != nil {
		return []string{}, err
	}

	return ttyUSBs, nil
}

func (clerk *PeripheralDeviceClerk) addDevice(dev *PeripheralDevice) {
	clerk.Devices[string(dev.Name[:])] = dev
}

func (clerk *PeripheralDeviceClerk) FindDevices() error {
	ports, err := clerk.listPorts()
	if err != nil {
		return err
	}

	for _, p := range ports {
		newDevice := NewPeripheralDevice(p)

		err := newDevice.Probe()
		if err != nil {
			fmt.Println(p, err.Error())
			continue
		}

		clerk.addDevice(newDevice)
		fmt.Println(p, string(newDevice.Name[:]))
	}

	// go clerk.clerkBackgroundJob()

	return nil
}

func (clerk *PeripheralDeviceClerk) ListDevices() error {
	for _, d := range clerk.Devices {
		fmt.Printf("[%d] %s\n", d.ID, d.Name)
	}

	return nil
}

// func (c *PeripheralDeviceClerk) clerkBackgroundJob() {
// 	for {
// 		c.mu.RLock()
// 		for _, d := range c.Devices {
// 			// Check the state of the device
// 			if d.CommState == CommIdle {
// 			}
// 		}
// 		c.mu.Unlock()
// 	}
// }
