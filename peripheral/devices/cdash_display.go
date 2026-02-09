package devices

import (
	"fmt"
	"os"
	"strconv"
	"unicode/utf8"

	helper "esdi/helpers"
	"esdi/peripheral/types"

	"golang.org/x/term"
)

const (
	newWindowCMDID     types.Command = 3
	destroyWindowCMDID types.Command = 4
	moveWindowCMDID    types.Command = 5
	newLayoutCMDID     types.Command = 6
)

var (
	DefaultDecorations = UIDecorations{
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

type LayoutTree struct {
	Windows map[int]UIWindow
}

func (l *LayoutTree) AddWindow(idx int, w UIWindow) {
	l.Windows[idx] = w
}

func NewLayoutTree() *LayoutTree {
	return &LayoutTree{
		Windows: make(map[int]UIWindow),
	}
}

type CDashState struct {
	Layout *LayoutTree
}

// Well fuck it really, this will make do for now really lmao
var (
	state *CDashState
)

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
		"move-window": {
			Identifier: moveWindowCMDID,
			Name:       "move-window",
			Desc:       "moves a window by its ID",
			ArgCheck:   moveWindowArgCheck,
			Fn:         moveWindow,
		},
		"new-layout": {
			Identifier: newLayoutCMDID,
			Name:       "new-layout",
			Desc:       "creates a new layout",
			ArgCheck:   nil,
			Fn:         nil,
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
		Decor: DefaultDecorations,
		Title: helper.B32(args[4]),
	}

	bytes, err := helper.StructToBytes(data)
	if err != nil {
		return 0, []byte{}, err
	}

	// All went well so far so we can update the state
	if state == nil {
		state = &CDashState{Layout: NewLayoutTree()}
	}

	state.Layout.AddWindow(0, data)

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

func moveWindowArgCheck(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("wrong parameters, got %d, want %d. "+
			"Command asks for: winID", len(args), 1)
	}

	return nil
}

func moveWindow(dCMD *DeviceCMD, args []string) (types.Command, []byte, error) {
	// Create an interactive mode to move the window or whatever
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return dCMD.GetIdentifier(), []byte{}, err
	}
	defer term.Restore(fd, oldState)

	buf := make([]byte, 8)
	// var buffer []byte

	for {
		var stop = false
		// We read the input on stdin and capture all types of inputs!
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return dCMD.GetIdentifier(), []byte{}, err
		}

		data := buf[:n]

		if len(data) <= 0 {
			continue
		}

		if data[0] == 0x1b {
			continue
		}

		r, _ := utf8.DecodeRune(data)
		if r == utf8.RuneError {
			return dCMD.GetIdentifier(), []byte{}, err
		}

		switch r {
		case 'q':
			stop = true
		case 'h':
			fmt.Printf("←")
		case 'l':
			fmt.Printf("→")
		case 'k':
			fmt.Printf("↑")
		case 'j':
			fmt.Printf("↓")
		}

		if stop {
			break
		}
	}

	return dCMD.GetIdentifier(), []byte{}, fmt.Errorf("do nothing")
}
