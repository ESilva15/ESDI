// Package devices is a peripheral factory
package devices

import (
	"errors"

	"esdi/devices/cdashdisplay"
	"esdi/devices/uidevice"
	"esdi/peripheral"
)

var ErrInvalidDevice = errors.New("invalid device")

type Device struct {
	Name     string
	Discover func() (peripheral.Peripheral, error)
	// DefaultSetup func(peripheral.Peripheral) error
}

var List map[string]*Device = map[string]*Device{
	uidevice.NAME: {
		Name:     uidevice.NAME,
		Discover: UIDeviceDiscover,
	},
	cdashdisplay.NAME: {
		Name:     cdashdisplay.NAME,
		Discover: CDashDisplayDiscover,
	},
}

func UIDeviceDiscover() (peripheral.Peripheral, error) {
	uidev, err := uidevice.NewUIDevice()
	if err != nil {
		return nil, err
	}

	return uidev, nil
}

// func UIDeviceSetup(peripheral peripheral.Peripheral) error {
// 	return nil
// }

func CDashDisplayDiscover() (peripheral.Peripheral, error) {
	// Create a cdashdisplay
	display, err := cdashdisplay.NewCDashDisplay()
	if err != nil {
		return nil, err
	}

	return display, nil
}

// func CDashDisplaySetup(peripheral peripheral.Peripheral) error {
// 	cdash, ok := peripheral.(*cdashdisplay.CDashDisplay)
// 	if !ok {
// 		return ErrInvalidDevice
// 	}
//
// 	err := cdash.LoadLayout("layout.yaml")
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
