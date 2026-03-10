package services

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/peripheral"
	"esdi/telemetry"
	"fmt"
	"log/slog"
	"sync/atomic"
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

func (cds *CDashService) CreateWindow(
	win *cdashdisplay.DesktopUIWindow,
) (*cdashdisplay.DesktopUIWindow, error) {
	updatedWindow, err := cds.CDash.CreateWindow(win)
	if err != nil {
		return nil, err
	}

	return updatedWindow, nil
}

func (cds *CDashService) LoadLayout(layoutPath string) error {
	return cds.CDash.LoadLayout(layoutPath)
}

func (cds *CDashService) SaveLayout(layoutPath string) error {
	return cds.CDash.SaveLayout(layoutPath)
}

func (cds *CDashService) UpdateWindow(win *cdashdisplay.DesktopUIWindow) error {
	cds.Messages <- fmt.Sprintf("Updating a window:\n%+v\n", win)
	return cds.CDash.UpdateWindow(win)
}

func (cds *CDashService) DeleteWindow(idx int16) error {
	return cds.CDash.DestroyWindow(idx)
}

func (cds *CDashService) ResizeWindow(idx int16, vec *helper.Vector) error {
	err := cds.CDash.ResizeWindow(idx, vec)
	if err != nil {
		return err
	}

	return nil
}

func (cds *CDashService) MoveWindow(idx int16, vec *helper.Vector) error {
	err := cds.CDash.MoveWindow(idx, vec)
	if err != nil {
		return err
	}

	return nil
}

func (cds *CDashService) StreamData(stream <-chan telemetry.TelemetryData) {
	var isSending atomic.Bool

	go func() {

		for msg := range stream {
			if isSending.Load() {
				continue
			}

			isSending.Store(true)

			cds.CDash.SendData(&msg)
			isSending.Store(false)
		}
	}()
}
