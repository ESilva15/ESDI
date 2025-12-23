// Package peripheral will handle communicating with peripherals
package peripheral

import (
	"esdi/peripheral/devices"
	"fmt"
	"path/filepath"
)

type PeripheralType string

const (
	DisplayPeripheral PeripheralType = "display"
)

type PeripheralDeviceClerk struct {
	// mu      sync.RWMutex
	Devices map[uint8]*PeripheralDevice
}

func NewPeripheralDeviceClerk() *PeripheralDeviceClerk {
	return &PeripheralDeviceClerk{
		Devices: make(map[uint8]*PeripheralDevice),
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
	clerk.Devices[dev.ID] = dev
}

func (clerk *PeripheralDeviceClerk) FindDeviceAPI(ID uint8) *devices.Device {
	for _, d := range devices.DeviceMap {
		if d.ID == ID {
			return &d
		}
	}

	return nil
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

		// Look for the device on our device list - we need its API
		newDevice.DeviceAPI = clerk.FindDeviceAPI(newDevice.ID)
		if newDevice.DeviceAPI == nil {
			fmt.Printf("Failed to acquire API for device: [%2d] %s\n",
				newDevice.ID, newDevice.Name)
		}

		clerk.addDevice(newDevice)
		fmt.Println(p, string(newDevice.Name[:]))
	}

	// go clerk.clerkBackgroundJob()

	return nil
}

func (clerk *PeripheralDeviceClerk) ListDevices() error {
	for _, d := range clerk.Devices {
		fmt.Printf("[%2d] %s\n", d.ID, d.Name)
	}

	return nil
}

func (clerk *PeripheralDeviceClerk) getDevice(ID uint8) *PeripheralDevice {
	if _, ok := clerk.Devices[ID]; !ok {
		return nil
	}

	return clerk.Devices[ID]
}

func (clerk *PeripheralDeviceClerk) ListDeviceAPI(ID uint8) error {
	dev := clerk.getDevice(ID)
	if dev == nil {
		return fmt.Errorf("device with ID %d not found", ID)
	}

	for _, f := range dev.DeviceAPI.API {
		fmt.Printf("%s\t%s\n", f.Name, f.Desc)
	}

	return nil
}

func (clerk *PeripheralDeviceClerk) RunDeviceFunction(ID uint8, f string, args []string) error {
	dev := clerk.getDevice(ID)
	if dev == nil {
		return fmt.Errorf("device with ID %d not found", ID)
	}
	devAPI := dev.DeviceAPI

	cmd := devAPI.HasFunction(f)
	if cmd == nil {
		return fmt.Errorf("requested function %s doesnt exist", f)
	}

	// Execute the function
	command, payload, err := cmd.Fn(cmd, args)
	if err != nil {
		return err
	}

	err = dev.SendCommand(command, payload)
	if err != nil {
		return err
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
