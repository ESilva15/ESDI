// Package cdashdisplay will handle the USB connections
package cdashdisplay

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"

	helper "esdi/helpers"
	"esdi/peripheral"
	"esdi/peripheral/communication"
	"esdi/peripheral/communication/packets"
	"esdi/peripheral/types"
	"esdi/telemetry"

	"gopkg.in/yaml.v3"
)

// I have to move this to some kind of configuration place
const (
	layoutsDir = "./layouts/"
)

const (
	newWindowCMDID        types.Command = 3
	destroyWindowCMDID    types.Command = 4
	updateWindowDimsCMDID types.Command = 5
	updateWindowCMDID     types.Command = 6 // Change this to a move cmd instead
	sendDataCMDID         types.Command = 7
	newLayoutCMDID        types.Command = 8
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

type CDashState struct {
	Layout *LayoutTree
}

func NewCDashState() *CDashState {
	return &CDashState{
		Layout: NewLayoutTree(),
	}
}

type CDashDisplay struct {
	WT                          *communication.WalkieTalkie
	State                       *CDashState
	fieldToWindows              map[telemetry.FieldID][]int16
	bufPool                     sync.Pool
	failedSends                 int
	FailedSendsConsecutiveLimit int
}

// Connect will try to find and connect to the CDashDisplay
func NewCDashDisplay() (*CDashDisplay, error) {
	// Look for the port
	p, err := findDisplayPort()
	if err != nil {
		slog.Info("failed to find cdashdisplay port", "reason", err.Error())
		return nil, err
	}

	return &CDashDisplay{
		WT:             p,
		State:          NewCDashState(),
		fieldToWindows: make(map[telemetry.FieldID][]int16),
		bufPool: sync.Pool{
			New: func() any {
				b := make([]byte, 0, telemetry.MaxFields*8)
				return &b
			},
		},
		failedSends:                 0,
		FailedSendsConsecutiveLimit: 5,
	}, nil
}

func (d *CDashDisplay) SendCommand() {
}

func (d *CDashDisplay) RegisterFieldMapping(fieldID telemetry.FieldID, winID int16) {
	d.fieldToWindows[fieldID] = append(d.fieldToWindows[fieldID], winID)
}

func (d *CDashDisplay) UnregisterFieldMapping(winID int16) {
	for fieldID, windows := range d.fieldToWindows {
		updated := windows[:0]
		for _, w := range windows {
			if w != winID {
				updated = append(updated, w)
			}

			if len(updated) == 0 {
				delete(d.fieldToWindows, fieldID)
			} else {
				d.fieldToWindows[fieldID] = updated
			}
		}
	}
}

func (d *CDashDisplay) CreateWindow(win *DesktopUIWindow) (*DesktopUIWindow, error) {
	bytes, err := helper.StructToBytes(win.UIWindow)
	if err != nil {
		return nil, err
	}

	// Send the command
	var wID packets.NewWindowID
	err = d.WT.SendCommand(newWindowCMDID, bytes, &wID)
	if err != nil {
		return nil, err
	}

	win.UIData.IDX = wID.ID

	slog.Info(fmt.Sprintf("Recived ID message: %v", wID))

	d.State.Layout.AddWindow(win)
	if fieldID, ok := telemetry.GetFieldID(win.UIData.TelemetryField); ok {
		d.RegisterFieldMapping(fieldID, win.UIData.IDX)
	}

	return win, nil
}

func (d *CDashDisplay) UpdateWindow(win *DesktopUIWindow) error {
	data := UIWindowUpdatePacket{
		WinID:  win.UIData.IDX,
		Window: win.UIWindow,
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
	slog.Debug(fmt.Sprintf("PreUpdate ID:  %p", win))
	d.State.Layout.Windows[win.UIData.IDX] = win
	slog.Debug(fmt.Sprintf("PostUpdate ID: %p", win))
	// Yeah, same address as suspected
	// I can't think about it right now. I'll think about that tomorrow

	// TODO: need to update the field mappings here!

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

	// NOTE: add this
	d.UnregisterFieldMapping(wID)
	d.State.Layout.RemoveWindow(wID)

	return nil
}

func (d *CDashDisplay) updateWindowDimensions(win *UIWindow, packet UpdateDimsPacket) error {
	slog.Debug(fmt.Sprintf("UPDATE: %v", packet))

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
	slog.Debug(fmt.Sprintf("cur dims: %v", win.Dims))
	win.Dims = packet.Dims
	slog.Debug(fmt.Sprintf("new dims: %v", win.Dims))

	return nil
}

func (d *CDashDisplay) MoveWindow(wID int16, delta *helper.Vector) error {
	// Get the window from the layout
	window, ok := d.State.Layout.Windows[wID]
	if !ok {
		return fmt.Errorf("window with ID '%d' doesn't exist", wID)
	}

	// Now we will create new dimensions
	newDimensions := window.Dims

	newDimensions.X0 += delta.DX
	newDimensions.Y0 += delta.DY

	packet := UpdateDimsPacket{
		ID:   wID,
		Dims: newDimensions,
	}

	err := d.updateWindowDimensions(&window.UIWindow, packet)
	if err != nil {
		return err
	}
	window.Dims = newDimensions

	return nil
}

func (d *CDashDisplay) ResizeWindow(wID int16, delta *helper.Vector) error {
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

	err := d.updateWindowDimensions(&window.UIWindow, packet)
	if err != nil {
		return err
	}
	window.Dims = newDimensions

	return nil
}

func (d *CDashDisplay) SaveLayout(outputPath string) error {
	file, err := os.OpenFile(path.Join(layoutsDir, outputPath),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(d.State.Layout)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (d *CDashDisplay) LoadLayout(layoutName string) error {
	data, err := os.ReadFile(path.Join(layoutsDir, layoutName))
	if err != nil {
		return err
	}

	layout := NewLayoutTree()
	err = yaml.Unmarshal(data, layout)
	if err != nil {
		return err
	}

	for _, w := range layout.Windows {
		_, err = d.CreateWindow(w)
		if err != nil {
			// NOTE: Add a way to handle multiple errors ?
			return err
		}
	}

	return nil
}

func (d *CDashDisplay) UnloadLayout() error {
	var err error
	for _, w := range d.State.Layout.Windows {
		slog.Debug(fmt.Sprintf("= Removing %d ==============================================",
			w.UIData.IDX))

		err = d.DestroyWindow(w.UIData.IDX)
		time.Sleep(75 * time.Millisecond)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to destroy window: %+v", err))
			// NOTE: Add a way to handle multiple errors ?
			return err
		}

		slog.Debug(fmt.Sprintf("= Removing %d ==============================================",
			w.UIData.IDX))
	}

	return nil
}

func (d *CDashDisplay) SendData(data *telemetry.TelemetryData) error {
	packet := d.encodePacket(data)

	bytes, err := helper.StructToBytes(packet)
	if err != nil {
		return peripheral.ErrFailureToPackData
	}

	curStr := ""
	byteCount := 0
	for _, byte := range bytes {
		byteCount++
		curStr += fmt.Sprintf("%02x ", byte)

		if byteCount == 8 {
			curStr = ""
			byteCount = 0
		}
	}

	// var ack packets.AckPacket
	err = d.WT.SendCommand(sendDataCMDID, bytes, nil)
	if err != nil && err != io.EOF {
		if d.failedSends == d.FailedSendsConsecutiveLimit {
			return peripheral.ErrDeviceTimedOut
		}
		d.failedSends++
	}

	return nil
}
