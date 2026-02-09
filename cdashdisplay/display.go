// Package cdashdisplay will handle the USB connections
package cdashdisplay

import (
	helper "esdi/helpers"
	"esdi/peripheral/communication"
	"esdi/peripheral/communication/packets"
	"esdi/peripheral/types"
	"fmt"
	"log/slog"
)

var pLogger *slog.Logger

func SetLogger(l *slog.Logger) {
	pLogger = l
}

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
	Windows map[int16]UIWindow
}

func NewLayoutTree() *LayoutTree {
	return &LayoutTree{
		Windows: make(map[int16]UIWindow),
	}
}

func (l *LayoutTree) AddWindow(idx int16, w UIWindow) {
	l.Windows[idx] = w
}

type CDashState struct {
	Layout *LayoutTree
}

func NewCDashState() *CDashState {
	return &CDashState{
		Layout: NewLayoutTree(),
	}
}

type CDashDisplay struct {
	WT    *communication.WalkieTalkie
	State *CDashState
}

func NewCDashDisplay() (*CDashDisplay, error) {
	// Look for the port
	p, err := findDisplayPort()
	if err != nil {
		pLogger.Info("failed to find cdashdisplay port: %s", err.Error())
		return nil, err
	}

	return &CDashDisplay{
		WT:    p,
		State: NewCDashState(),
	}, nil
}

func (d *CDashDisplay) CreateWindow(win UIWindow) (int16, error) {
	bytes, err := helper.StructToBytes(win)
	if err != nil {
		return -1, err
	}

	// Send the command
	var wID packets.NewWindowID
	err = d.WT.SendCommand(newWindowCMDID, bytes, &wID)
	if err != nil {
		return -1, err
	}

	pLogger.Info(fmt.Sprintf("Recived ID message: %v", wID))

	d.State.Layout.AddWindow(wID.ID, win)

	return wID.ID, nil
}

func (d *CDashDisplay) DestroyWindow(wID int16) error {
	type UIWindowDestructPacket struct {
		WinID int16
	}

	packet := UIWindowDestructPacket{
		WinID: wID,
	}

	bytes, err := helper.StructToBytes(packet)
	if err != nil {
		return err
	}

	var ack packets.AckPacket
	err = d.WT.SendCommand(destroyWindowCMDID, bytes, &ack)
	if err != nil {
		return err
	}

	// NODE: add this
	// d.State.Layout.RemoveWindow(wID)

	return nil
}
