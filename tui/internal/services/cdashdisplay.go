package services

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/peripheral"
	"esdi/tui/internal/models"
	"log/slog"
)

type CDashService struct {
	Logger *slog.Logger
	CDash  *cdashdisplay.CDashDisplay
	// iRacingTelemetry *IRacingService
	DevClerk *peripheral.PeripheralDeviceClerk
	Messages chan string
}

func NewCDashService(logger *slog.Logger) *CDashService {
	sharedChannel := make(chan string, 10)
	return &CDashService{
		Logger:   logger,
		CDash:    nil,
		DevClerk: peripheral.NewPeripheralDeviceClerk(),
		Messages: sharedChannel,
		// iRacingTelemetry: NewIRacingService(sharedChannel),
	}
}

func (cds *CDashService) FindDevice() {
	cds.Messages <- "looking for cdash display...\n"
	cds.Logger.Info("Looking for CDashDisplay")

	cdashdisplay.SetLogger(cds.Logger.With("[device]", "cdashdisplay"))

	display, err := cdashdisplay.NewCDashDisplay()
	if err != nil {
		cds.Logger.Info("didn't find cdashdisplay")
		cds.Messages <- "didn't find cdash display\n"
		return
	}

	cds.CDash = display
	cds.Logger.Info("found cdashdisplay on: " + display.WT.Cfg.Name)
	cds.Messages <- "found cdashdisplay on: " + display.WT.Cfg.Name + "\n"
}

func (cds *CDashService) CreateWindow(win *cdashdisplay.UIWindow) (int16, error) {
	wID, err := cds.CDash.CreateWindow(*win)
	if err != nil {
		return -1, err
	}

	return wID, nil
}

func (cds *CDashService) LoadLayout(layoutPath string) error {
	return cds.CDash.LoadLayout(layoutPath)
}

func (cds *CDashService) SaveLayout(layoutPath string) error {
	return cds.CDash.SaveLayout(layoutPath)
}

func (cds *CDashService) UpdateWindow(idx int16, win *cdashdisplay.UIWindow) error {
	return cds.CDash.UpdateWindow(idx, win)
}

func (cds *CDashService) DeleteWindow(idx int16) error {
	return cds.CDash.DestroyWindow(idx)
}

func (cds *CDashService) ResizeWindow(win *models.UIWindow, vec *helper.Vector) error {
	err := cds.CDash.ResizeWindow(win.IDX, vec)
	if err != nil {
		return err
	}

	// Update the UI window data
	win.Window = *cds.CDash.State.Layout.Windows[win.IDX]

	return nil
}

func (cds *CDashService) MoveWindow(win *models.UIWindow, vec *helper.Vector) error {
	err := cds.CDash.MoveWindow(win.IDX, vec)
	if err != nil {
		return err
	}

	// Update the UI window data
	win.Window = *cds.CDash.State.Layout.Windows[win.IDX]

	return nil
}

// func (cds *CDashService) StartStream() {
// 	// We need to get a channel from the telemetry agent to listen to and
// 	// propogate the data to the clients
// 	dataStream := cds.iRacingTelemetry.StartStream()
//
// 	// Custom types to be able to actually send data to the device
// 	type CDashDisplayUIWindowStreamData struct {
// 		IDX  int16
// 		Data [32]byte
// 	}
//
// 	type CDashDisplayStreamData struct {
// 		Data []CDashDisplayUIWindowStreamData
// 	}
//
// 	// We will need a mechanism to kill this channel too I reckon
// 	go func() {
// 		for data := range dataStream {
// 			// We have to send this data to the device
// 			cds.Messages <- fmt.Sprintf("%d %d %d\n", data.Speed[:], data.Gear[:], data.RPM[:])
// 		}
// 	}()
// }

// func (cds *CDashService) GetStream() <-chan string {
// 	return cds.iRacingTelemetry.GetStream()
// }
//
// func (cds *CDashService) StopStream() {
// 	cds.iRacingTelemetry.StopStream()
// }
