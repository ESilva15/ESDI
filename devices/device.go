// Package devices is a peripheral factory
package devices

import (
	"esdi/devices/cdashdisplay"
	"esdi/devices/uidevice"
	"esdi/peripheral"
)

type Device struct {
	Name     string
	Discover func() (peripheral.Peripheral, error)
}

var List map[string]Device = map[string]Device{
	uidevice.NAME: {
		Name:     uidevice.NAME,
		Discover: DiscoverUIDevice,
	},
	cdashdisplay.NAME: {
		Name:     cdashdisplay.NAME,
		Discover: DiscoverCDashDisplay,
	},
}

func DiscoverUIDevice() (peripheral.Peripheral, error) {
	uidev, err := uidevice.NewUIDevice()
	if err != nil {
		return nil, err
	}

	return uidev, nil
}

func DiscoverCDashDisplay() (peripheral.Peripheral, error) {
	// Create a cdashdisplay
	display, err := cdashdisplay.NewCDashDisplay()
	if err != nil {
		return nil, err
	}

	return display, nil
}
