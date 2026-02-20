// Package cdashdisplay will handle the USB connections
package cdashdisplay

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"

	helper "esdi/helpers"
	"esdi/peripheral/communication"
	"esdi/peripheral/communication/packets"
	"esdi/peripheral/types"

	"gopkg.in/yaml.v3"
)

var pLogger *slog.Logger

func SetLogger(l *slog.Logger) {
	pLogger = l
}

// I have to move this to some kind of configuration place
const (
	layoutsDir = "./layouts/"
)

const (
	newWindowCMDID        types.Command = 3
	destroyWindowCMDID    types.Command = 4
	updateWindowDimsCMDID types.Command = 5
	updateWindowCMDID     types.Command = 6 // Change this to a move cmd instead
	newLayoutCMDID        types.Command = 7
)

const (
	MoveLeft = iota
	MoveDown
	MoveUp
	MoveRight
)

var DefaultDecorations = UIDecorations{
	HasBorder:    1,
	BGColour:     0x1041, // Look in the eslabsCurses library for these colours
	FGColour:     0xffff,
	TitleColour:  0xffff,
	BorderColour: 0xf800,
	TitleSize:    2,
	TextSize:     4,
	Padding:      0x00,
}

type UpdateDimsPacket struct {
	ID   int16
	Dims UIDimensions
}

type FString32 [32]byte

func (s FString32) String() string {
	n := bytes.IndexByte(s[:], 0)
	if n == -1 {
		n = len(s)
	}
	return string(s[:n])
}

func (s FString32) MarshalYAML() (any, error) {
	// trim trailing zero bytes
	n := bytes.IndexByte(s[:], 0)
	if n == -1 {
		n = len(s)
	}
	return string(s[:n]), nil
}

func (s *FString32) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected YAML scalar for FixedString32")
	}

	b := []byte(value.Value)

	if len(b) > len(s) {
		return fmt.Errorf("string too long (max %d bytes)", len(s))
	}

	// zero-fill first
	for i := range s {
		s[i] = 0
	}

	copy(s[:], b)
	return nil
}

type LayoutTree struct {
	Windows map[int16]*UIWindow `yaml:"Windows"`
}

func NewLayoutTree() *LayoutTree {
	return &LayoutTree{
		Windows: make(map[int16]*UIWindow),
	}
}

func (l *LayoutTree) AddWindow(idx int16, w UIWindow) {
	pLogger.Debug(fmt.Sprintf("adding window '%d' - %v", idx, w))
	l.Windows[idx] = &w
	pLogger.Debug(fmt.Sprintf("new map - %v", l.Windows))
}

func (l *LayoutTree) RemoveWindow(idx int16) {
	pLogger.Debug(fmt.Sprintf("removing window '%d'", idx))
	delete(l.Windows, idx)
	pLogger.Debug(fmt.Sprintf("new map - %v", l.Windows))
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

func (d *CDashDisplay) SendCommand() {
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

func (d *CDashDisplay) UpdateWindow(wID int16, win *UIWindow) error {
	data := UIWindowUpdatePacket{
		WinID:  wID,
		Window: *win,
	}

	bytes, err := helper.StructToBytes(data)
	if err != nil {
		return err
	}

	err = d.WT.SendCommand(updateWindowCMDID, bytes, nil)
	if err != nil {
		return err
	}

	// NOTE: need to check the flow of this because:
	// windows has a pointer to a UIWindow stored
	// I get it and update it in the controller
	// I send the pointer here
	// -> it should be the same pointer then right?
	pLogger.Debug(fmt.Sprintf("PreUpdate ID:  %p", win))
	d.State.Layout.Windows[wID] = win
	pLogger.Debug(fmt.Sprintf("PostUpdate ID: %p", win))
	// Yeah, same address as suspected
	// I can't think about it right now. I'll think about that tomorrow

	return nil
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

	err = d.WT.SendCommand(destroyWindowCMDID, bytes, nil)
	if err != nil {
		return err
	}

	// NODE: add this
	d.State.Layout.RemoveWindow(wID)

	return nil
}

func (d *CDashDisplay) updateWindowDimensions(win *UIWindow, packet UpdateDimsPacket) error {
	pLogger.Debug(fmt.Sprintf("UPDATE: %v", packet))

	bytes, err := helper.StructToBytes(packet)
	if err != nil {
		return err
	}

	// var ack packets.AckPacket
	err = d.WT.SendCommand(updateWindowDimsCMDID, bytes, nil)
	if err != nil && err != io.EOF {
		return err
	}

	// Nothing bad happened afaik
	pLogger.Debug(fmt.Sprintf("cur dims: %v", win.Dims))
	win.Dims = packet.Dims
	pLogger.Debug(fmt.Sprintf("new dims: %v", win.Dims))

	return nil
}

func (d *CDashDisplay) MoveWindow(wID int16, delta helper.Vector) (UIDimensions, error) {
	// Get the window from the layout
	window, ok := d.State.Layout.Windows[wID]
	if !ok {
		return UIDimensions{}, fmt.Errorf("window with ID '%d' doesn't exist", wID)
	}

	// Now we will create new dimensions
	newDimensions := window.Dims

	newDimensions.X0 += delta.DX
	newDimensions.Y0 += delta.DY

	packet := UpdateDimsPacket{
		ID:   wID,
		Dims: newDimensions,
	}

	return newDimensions, d.updateWindowDimensions(window, packet)
}

func (d *CDashDisplay) ResizeWindow(wID int16, delta helper.Vector) error {
	// Get the window from the layout
	window, ok := d.State.Layout.Windows[wID]
	if !ok {
		return fmt.Errorf("window with ID '%d' doesn't exist", wID)
	}

	// Now we will create new dimensions
	newDimensions := window.Dims

	if delta.DX > 0 {
		newDimensions.Width += delta.DX
	} else {
		newDimensions.Width -= delta.DX
		newDimensions.X0 -= 2 * delta.DX
	}

	if delta.DY > 0 {
		newDimensions.Height += delta.DY
	} else {
		newDimensions.Height -= delta.DY
		newDimensions.Y0 -= 2 * delta.DY
	}

	packet := UpdateDimsPacket{
		ID:   wID,
		Dims: newDimensions,
	}

	return d.updateWindowDimensions(window, packet)
}

func (d *CDashDisplay) SaveLayout() {
	file, err := os.OpenFile(path.Join(layoutsDir, "layout.yaml"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic("failed to open logging file: " + err.Error())
	}

	data, err := yaml.Marshal(d.State.Layout)
	if err != nil {
		panic("failed to marshal layout data")
	}

	_, err = file.Write(data)
	if err != nil {
		panic("failed to write data to file")
	}
}

func (d *CDashDisplay) LoadLayout() {
	data, err := os.ReadFile(path.Join(layoutsDir, "layout.yaml"))
	if err != nil {
		panic(err.Error())
	}

	layout := NewLayoutTree()
	err = yaml.Unmarshal(data, layout)
	if err != nil {
		panic(err.Error())
	}

	for _, w := range layout.Windows {
		_, err = d.CreateWindow(*w)
		if err != nil {
			pLogger.Debug(fmt.Sprintf("Failed to create window: %v", w))
		}
	}
}
