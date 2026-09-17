// Package devices is a peripheral factory
package devices

import (
	"errors"

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
	// // Find CDashDisplay
	// {
	// 	ds.Messages <- "looking for " + cdashdisplay.Name + "...\n"
	// 	ds.Logger.Info("Looking for " + cdashdisplay.Name)
	//
	// 	cdashdisplay.SetLogger(ds.Logger.With("[device]", cdashdisplay.Name))
	//
	// 	// Create a cdashdisplay
	// 	display, err := cdashdisplay.Discover()
	// 	if err == nil {
	// 		ds.Devices[cdashdisplay.Name] = display
	// 		ds.Logger.Info("found " + cdashdisplay.Name + " on: " + display.WT.Cfg.Name)
	// 		ds.Messages <- "found " + cdashdisplay.Name + " on: " + display.WT.Cfg.Name + "\n"
	// 		return
	// 	}
	//
	// 	ds.Logger.Info("didn't find " + cdashdisplay.Name)
	// 	ds.Messages <- "didn't find " + cdashdisplay.Name + "\n"
	// 	// No CDashDisplay available for one reason or another, so we don't set the
	// 	// key
	// }
	return nil, errors.New("not implemented yet")
}
