package devices

import (
	"fmt"
	"strconv"

	helper "esdi/helpers"
	"esdi/peripheral/types"
)

const (
	newWindowCMDID     types.Command = 3
	destroyWindowCMDID types.Command = 4
)

var (
	defaultDecorations = UIDecorations{
		HasBorder:    1,
		BGColour:     0x1041, // Look in the eslabsCurses library for these colours
		FGColour:     0xffff,
		TitleColour:  0xffff,
		BorderColour: 0xf800,
		TitleSize:    2,
		TextSize:     4,
		Padding:      0x00,
	}
)

type UIDimensions struct {
	X0     uint16
	Y0     uint16
	Width  uint16
	Height uint16
}

type UIDecorations struct {
	BGColour     uint16
	FGColour     uint16
	TitleColour  uint16
	BorderColour uint16
	TitleSize    uint8
	TextSize     uint8
	HasBorder    uint8
	Padding      uint8
}

type UIWindow struct {
	Dims  UIDimensions
	Decor UIDecorations
	Title [32]byte
}

var CDashDisplay = Device{
	ID:   CDashDisplayDevID,
	Name: helper.B32("ESLabs CDashDisplay"),
	API: map[string]DeviceCMD{
		"new-window": {
			Identifier: newWindowCMDID,
			Name:       "new-window",
			Desc:       "Creates a new window - pass the window name and dimensions",
			ArgCheck:   createWindowArgCheck,
			Fn:         createWindow,
		},
		"destroy-window": {
			Identifier: destroyWindowCMDID,
			Name:       "destroy-window",
			Desc:       "Destroys a window by its ID",
			ArgCheck:   destroyWindowArgCheck,
			Fn:         destroyWindow,
		},
	},
}

func createWindowArgCheck(args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("wrong parameters, got %d, want %d. "+
			"Command asks for: x0 y0 width height title", len(args), 5)
	}

	return nil
}

func createWindow(dCMD *DeviceCMD, args []string) (types.Command, []byte, error) {
	// Parse the command
	// x0, y0, width, height, title # add other decorations later on
	// fmt.Printf("%s command called: %+v\n", dCMD.GetName(), args)

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
		Dims: UIDimensions{
			X0:     uint16(x0),
			Y0:     uint16(y0),
			Width:  uint16(width),
			Height: uint16(height),
		},
		Decor: defaultDecorations,
		Title: helper.B32(args[4]),
	}

	bytes, err := helper.StructToBytes(data)
	if err != nil {
		return 0, []byte{}, err
	}

	return dCMD.GetIdentifier(), bytes, nil
}

func destroyWindowArgCheck(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("wrong parameters, got %d, want %d. "+
			"Command asks for: winID", len(args), 1)
	}

	return nil
}

func destroyWindow(dCMD *DeviceCMD, args []string) (types.Command, []byte, error) {
	// Parse the command
	fmt.Printf("%s command called: %+v\n", dCMD.GetName(), args)

	type UIWindowDestructPacket struct {
		WinID int16
	}

	id, err := strconv.ParseInt(args[0], 10, 0)
	if err != nil {
		return 0, []byte{}, err
	}

	packet := UIWindowDestructPacket{
		WinID: int16(id),
	}

	bytes, err := helper.StructToBytes(packet)
	if err != nil {
		return 0, []byte{}, err
	}

	return dCMD.GetIdentifier(), bytes, nil
}
