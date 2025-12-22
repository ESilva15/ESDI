package devices

import (
	"fmt"
	"strconv"

	helper "esdi/helpers"
	"esdi/peripheral/types"
)

const (
	newWindowCMDID types.CommandID = iota
	destroyWindowCMDID
)

var CDashDisplay = Device{
	ID:   CDashDisplayDevID,
	Name: helper.B32("ESLabs CDashDisplay"),
	API: map[string]DeviceCMD{
		"new-window": {
			Identifier: newWindowCMDID,
			Name:       "new-window",
			Desc:       "Creates a new window - pass the window name and dimensions",
			Fn:         createWindow,
		},
		"destroy-window": {
			Identifier: destroyWindowCMDID,
			Name:       "destroy-window",
			Desc:       "Destroys a window by its ID",
			Fn:         destroyWindow,
		},
	},
}

func createWindow(dCMD *DeviceCMD, args []string) (types.CommandID, []byte, error) {
	// Parse the command
	// x0, y0, width, height, title # add other decorations later on
	// fmt.Printf("%s command called: %+v\n", dCMD.GetName(), args)

	type UIWindow struct {
		X0     uint16
		Y0     uint16
		Width  uint16
		Height uint16
		Title  [32]byte
	}

	if len(args) != 5 {
		return 0, []byte{}, fmt.Errorf("wrong parameters, got %d, want %d. "+
			"Command asks for: x0 y0 width height title", len(args), 5)
	}

	x0, err := strconv.ParseInt(args[0], 10, 0)
	if err != nil {
		return 0, []byte{}, err
	}
	y0, err := strconv.ParseInt(args[1], 10, 0)
	if err != nil {
		return 0, []byte{}, err
	}
	width, err := strconv.ParseInt(args[2], 10, 0)
	if err != nil {
		return 0, []byte{}, err
	}
	height, err := strconv.ParseInt(args[3], 10, 0)
	if err != nil {
		return 0, []byte{}, err
	}

	data := UIWindow{
		X0:     uint16(x0),
		Y0:     uint16(y0),
		Width:  uint16(width),
		Height: uint16(height),
		Title:  helper.B32(args[4]),
	}

	// fmt.Printf("Data: %+v\n", data)

	bytes, err := helper.StructToBytes(data)
	if err != nil {
		return 0, []byte{}, err
	}

	return dCMD.GetIdentifier(), bytes, nil
}

func destroyWindow(dCMD *DeviceCMD, args []string) (types.CommandID, []byte, error) {
	// Parse the command
	fmt.Printf("%s command called: %+v\n", dCMD.GetName(), args)

	return dCMD.GetIdentifier(), []byte{}, nil
}
