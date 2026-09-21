package services

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"esdi/devices"
	"esdi/peripheral"
)

var (
	ErrPeripheralAlreadyRegistered = errors.New("peripheral is already registered")
	ErrNoSuchDevice                = errors.New("device doesn't exist")
	ErrDeviceIsNotConnected        = errors.New("device isn't connected")
	ErrFailedToSetupPeripheral     = errors.New("peripheral setup failed")
)

type DeviceState = uint8

const (
	DeviceTimedOut uint8 = iota
	DeviceIsConnected
	DeviceIsDisconnected
	DeviceReconnected
)

type PeripheralState struct {
	device     *devices.Device
	Peripheral peripheral.Peripheral
	State      DeviceState
	Setup      bool
}

func NewPeripheralState(
	dev *devices.Device,
	peripheral peripheral.Peripheral,
	state DeviceState,
) *PeripheralState {
	perState := PeripheralState{
		device:     dev,
		Peripheral: peripheral,
		State:      state,
		Setup:      false,
	}

	return &perState
}

type PeripheralStateStore struct {
	Logger *slog.Logger
	mu     sync.RWMutex
	store  map[string]*PeripheralState
	// Messaging for UI and stuff
	Messages chan string
	// Callbacks
	telemetryProvider func() (string, error)
}

func NewPeripheralStateStore(
	nLogger *slog.Logger,
	devList map[string]*devices.Device,
	msg chan string,
) *PeripheralStateStore {
	store := PeripheralStateStore{
		Logger:   nLogger,
		store:    make(map[string]*PeripheralState),
		Messages: msg,
	}

	for _, dev := range devList {
		store.AddDevice(dev)
	}

	return &store
}

func (pss *PeripheralStateStore) GetStates() map[string]*PeripheralState {
	pss.mu.RLock()
	defer pss.mu.RUnlock()

	return maps.Clone(pss.store)
}

func (pss *PeripheralStateStore) GetState(pname string) (*PeripheralState, error) {
	if !pss.DeviceExists(pname) {
		return nil, ErrNoSuchDevice
	}

	pss.mu.RLock()
	defer pss.mu.RUnlock()
	return pss.store[pname], nil
}

func (pss *PeripheralStateStore) GetPeripheral(pname string) (peripheral.Peripheral, error) {
	state, err := pss.GetState(pname)
	if err != nil {
		return nil, err
	}

	if state.State != DeviceIsConnected {
		return nil, ErrDeviceIsNotConnected
	}

	return state.Peripheral, nil
}

// AddDevice adds a new device for tracking
func (pss *PeripheralStateStore) AddDevice(dev *devices.Device) error {
	if pss.DeviceExists(dev.Name) {
		return ErrPeripheralAlreadyRegistered
	}

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[dev.Name] = NewPeripheralState(dev, nil, DeviceIsDisconnected)

	return nil
}

func (pss *PeripheralStateStore) UpdatePeripheralSetupState(pname string, nState bool) error {
	if !pss.DeviceExists(pname) {
		return ErrNoSuchDevice
	}

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].Setup = nState

	return nil
}

// DeviceExists returns whether the store is already tracking `pname`
func (pss *PeripheralStateStore) DeviceExists(pname string) bool {
	pss.mu.RLock()
	defer pss.mu.RUnlock()

	if _, ok := pss.store[pname]; ok {
		return true
	}

	return false
}

// DeleteDevice deletes `pname` from tracking
func (pss *PeripheralStateStore) DeleteDevice(pname string) error {
	if !pss.DeviceExists(pname) {
		return ErrNoSuchDevice
	}

	pss.mu.Lock()
	defer pss.mu.Unlock()

	delete(pss.store, pname)

	return nil
}

// "Events" [START] ------------------------------------------------------------
func (ds *DeviceService) onDeviceTimedOut(pname string) {
	// We need to deregister the device
	ds.PSS.setDeviceTimedOut(pname)
}

// "Events" [END] --------------------------------------------------------------

// Device State Handling [START] -----------------------------------------------
func (pss *PeripheralStateStore) setDeviceConnected(pname string, per peripheral.Peripheral) {
	pss.Logger.Info("found device", "device", pname)

	pss.mu.Lock()
	pss.store[pname].Peripheral = per
	pss.store[pname].State = DeviceIsConnected
	pss.mu.Unlock()

	pss.Messages <- fmt.Sprintf("Device successfuly connected: %s\n", pname)
	// pss.OnDeviceFound(pname)
}

func (pss *PeripheralStateStore) setDeviceTimedOut(pname string) {
	pss.Logger.Info("device timed out", "device", pname)

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].Peripheral = nil
	pss.store[pname].State = DeviceTimedOut
}

func (pss *PeripheralStateStore) setDeviceReconnected(pname string, per peripheral.Peripheral) {
	pss.Logger.Info("device reconnecting", "device", pname)

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].Peripheral = per
	pss.store[pname].State = DeviceReconnected
}

// Device State Handling [END] -------------------------------------------------

// Device Handling [START] -----------------------------------------------------
func (pss *PeripheralStateStore) discoverPeripheral(
	pname string,
	onDiscovery func(pName string, peripheral peripheral.Peripheral),
) error {
	state, err := pss.GetState(pname)
	if err != nil {
		pss.Logger.Error("Can't reconnect device", "device", pname, "error", err)
		return err
	}

	dev, err := state.device.Discover()
	if err != nil {
		return err
	}

	// Register the device we just found
	onDiscovery(pname, dev)

	return nil
}

func (pss *PeripheralStateStore) handleDeviceTimedOut(pname string) error {
	err := pss.discoverPeripheral(pname, pss.setDeviceReconnected)
	if err != nil {
		// Log something
		return err
	}

	return nil
}

func (pss *PeripheralStateStore) handleDeviceReconnected(pname string) error {
	// The device has reconnected, but we must set its state again
	state, err := pss.GetState(pname)
	if err != nil {
		return err
	}

	// Update the peripheral state
	pss.setDeviceConnected(pname, state.Peripheral)
	pss.UpdatePeripheralSetupState(pname, false)

	return nil
}

// handleDeviceConnected will handle the device setup after it connects
// NOTE: should this be a state after Connected?
// Connected -> Unconfigured -> Configured I believe this would work nicely
// THIS IS A TODO ↑↑↑↑↑↑
func (pss *PeripheralStateStore) handleDeviceConnected(pname string) error {
	// Things to do once the device is connected
	// 1. Setup
	state, err := pss.GetState(pname)
	if err != nil {
		// We need to log something here or something
		return err
	}

	if !state.Setup {
		err = pss.setupPeripheral(state)
	}

	return nil
}

func (pss *PeripheralStateStore) setupPeripheral(state *PeripheralState) error {
	provider, err := pss.telemetryProvider()
	if err != nil {
		return err
	}

	err = state.Peripheral.Setup(provider)
	if err != nil {
		return err
	}

	err = pss.UpdatePeripheralSetupState(state.device.Name, true)
	if err != nil {
		return err
	}

	return nil
}

func (pss *PeripheralStateStore) HandleDeviceState() {
	peripherals := pss.GetStates()

	for pName, pState := range peripherals {
		switch pState.State {
		case DeviceIsDisconnected:
			pss.Logger.Debug("looking for device", "name", pName)
			err := pss.discoverPeripheral(pName, pss.setDeviceConnected)
			if err != nil {
				// Log something
				continue
			}
		case DeviceIsConnected:
			// Need to check if its streaming, if its not streaming than we have to do a healthcheck
			pss.Logger.Debug("Device is connected. Normal", "device", pName)
			pss.handleDeviceConnected(pName)
		case DeviceReconnected:
			// If the device has reconnected we need to reset the device and then set it as connected
			pss.Logger.Debug("Device has reconnected. Clearing up state", "device", pName)
			err := pss.handleDeviceReconnected(pName)
			if err != nil {
				pss.Logger.Error("device reconnection handler failed", "error", err)
				continue
			}
		case DeviceTimedOut:
			// If the device has timed out we need to re-discover it or something
			pss.Logger.Debug("Device is timed out. Attempting to recconect", "device", pName)
			err := pss.handleDeviceTimedOut(pName)
			if err != nil {
				pss.Logger.Error("device timing out handler failed", "error", err)
				continue
			}
		}
	}
}

// Device Handling [END] -------------------------------------------------------

// FindDevices is a routine that goes over the devices in the PeripheralStateStore
// and handles their state accordingly
func (ds *DeviceService) FindDevices() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctxDiscovery.Done():
			return
		case <-ticker.C:
			ds.PSS.HandleDeviceState()
		}
	}
}
