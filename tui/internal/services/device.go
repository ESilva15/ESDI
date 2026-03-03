package services

import (
	"esdi/cdashdisplay"
	"esdi/peripheral"
	"log/slog"
)

type CDashService struct {
	Logger   *slog.Logger
	CDash    *cdashdisplay.CDashDisplay
	DevClerk *peripheral.PeripheralDeviceClerk
	Messages chan string
}

func NewCDashService(logger *slog.Logger) *CDashService {
	return &CDashService{
		Logger:   logger,
		CDash:    nil,
		DevClerk: peripheral.NewPeripheralDeviceClerk(),
		Messages: make(chan string, 10),
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
