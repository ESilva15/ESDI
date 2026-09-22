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
	DeviceIsDisconnected
	DeviceReconnected
	DeviceIsDiscovering
	DeviceIsConnected
	DeviceIsUnconfigured
	DeviceIsConfiguring
	DeviceIsConfigured
	DeviceIsStreaming
)

func DeviceStateToStr(state DeviceState) string {
	switch state {
	case DeviceTimedOut:
		return "DeviceTimedOut"
	case DeviceIsDisconnected:
		return "DeviceIsDisconnected"
	case DeviceReconnected:
		return "DeviceReconnected"
	case DeviceIsDiscovering:
		return "DeviceIsDiscovering"
	case DeviceIsConnected:
		return "DeviceIsConnected"
	case DeviceIsUnconfigured:
		return "DeviceIsUnconfigured"
	case DeviceIsConfiguring:
		return "DeviceIsConfiguring"
	case DeviceIsConfigured:
		return "DeviceIsConfigured"
	case DeviceIsStreaming:
		return "DeviceIsStreaming"
	default:
		return "UnknownState"
	}
}

type PeripheralState struct {
	device     *devices.Device
	Peripheral peripheral.Peripheral
	State      DeviceState
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

	if state.State < DeviceIsConnected {
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
	ds.PSS.setDeviceTimedOut(pname)
}

func (pss *PeripheralStateStore) OnStartStream() {
	for _, state := range pss.GetStates() {
		if state.State == DeviceIsConfigured {
			pss.setDeviceIsStreaming(state.device.Name)
		}
	}
}

// "Events" [END] --------------------------------------------------------------

// Device State Handling [START] -----------------------------------------------
func (pss *PeripheralStateStore) setDeviceDisconnected(pname string) {
	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].State = DeviceIsDisconnected
}

func (pss *PeripheralStateStore) setDeviceIsDiscovering(pname string) {
	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].State = DeviceIsDiscovering
}

func (pss *PeripheralStateStore) setDeviceConnected(pname string, per peripheral.Peripheral) {
	pss.Logger.Info("found device", "device", pname)

	pss.mu.Lock()
	pss.store[pname].Peripheral = per
	pss.store[pname].State = DeviceIsConnected
	pss.mu.Unlock()

	pss.Messages <- fmt.Sprintf("Device successfuly connected: %s\n", pname)
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

func (pss *PeripheralStateStore) setDeviceUnconfigured(pname string) {
	pss.Logger.Info("device is connected but not configured", "device", pname)

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].State = DeviceIsUnconfigured
}

func (pss *PeripheralStateStore) setDeviceIsConfiguring(pname string) {
	pss.Logger.Info("device is configuring for new provider", "device", pname)

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].State = DeviceIsConfiguring
}

func (pss *PeripheralStateStore) setDeviceConfigured(pname string) {
	pss.Logger.Info("device is configured and ready for data", "device", pname)

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].State = DeviceIsConfigured
}

func (pss *PeripheralStateStore) setDeviceIsStreaming(pname string) {
	pss.Logger.Info("device is configured and ready for data", "device", pname)

	pss.mu.Lock()
	defer pss.mu.Unlock()
	pss.store[pname].State = DeviceIsStreaming
}

// Device State Handling [END] -------------------------------------------------

// Device Handling [START] -----------------------------------------------------
func (pss *PeripheralStateStore) discoverPeripheral(
	pname string,
	onDiscovery func(string, peripheral.Peripheral),
	onFailure func(string),
) {
	pss.Logger.Debug("looking for device", "name", pname)
	pss.setDeviceIsDiscovering(pname)
	pss.Messages <- "Discovering " + pname + "\n"

	go func() {
		state, err := pss.GetState(pname)
		if err != nil {
			pss.Logger.Error("Can't reconnect device", "device", pname, "error", err)
			onFailure(pname)
			return
		}

		dev, err := state.device.Discover()
		if err != nil {
			onFailure(pname)
			return
		}

		// Register the device we just found
		onDiscovery(pname, dev)
	}()
}

func (pss *PeripheralStateStore) configurePeripheral(
	pname string,
	onSuccess func(string),
	onFailure func(string),
) {
	go func() {
		state, err := pss.GetState(pname)
		if err != nil {
			onFailure(pname)
			return
		}

		provider, err := pss.telemetryProvider()
		if err != nil {
			onFailure(pname)
			return
		}

		err = state.Peripheral.Setup(provider)
		if err != nil {
			onFailure(pname)
			return
		}

		onSuccess(pname)
		pss.setDeviceConfigured(pname)
	}()
}

func (pss *PeripheralStateStore) handleDeviceTimedOut(pname string) error {
	pss.discoverPeripheral(pname, pss.setDeviceReconnected, pss.setDeviceTimedOut)
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

	return nil
}

// handleDeviceConnected will handle the device setup after it connects
// NOTE: should this be a state after Connected?
// Connected -> Unconfigured -> Configured I believe this would work nicely
// THIS IS A TODO ↑↑↑↑↑↑
func (pss *PeripheralStateStore) handleDeviceConnected(pname string) error {
	pss.setDeviceUnconfigured(pname)
	return nil
}

func (pss *PeripheralStateStore) handleDeviceIsUnconfigured(pname string) error {
	// Here we need to configure our device. If no error occurs its configured!
	_, err := pss.telemetryProvider()
	if err != nil {
		return err
	}

	pss.setDeviceIsConfiguring(pname)
	pss.configurePeripheral(pname, pss.setDeviceConfigured, pss.setDeviceUnconfigured)

	return nil
}

func (pss *PeripheralStateStore) handleDeviceIsConfigured(pname string) error {
	// Here we have to check wheter we are streaming or not. If we aren't streaming
	// then we ought to do a healthcheck on the peripheral
	return nil
}

func (pss *PeripheralStateStore) performHealthCheck(pname string, state *PeripheralState) bool {
	healthStatus := state.Peripheral.HealthCheck()
	if !healthStatus {
		pss.Messages <- "peripheral " + pname + " failed healthcheck"
		pss.setDeviceTimedOut(pname)
		return false
	}

	return true
}

func (pss *PeripheralStateStore) HandleDeviceState() {
	peripherals := pss.GetStates()

	for pName, pState := range peripherals {
		switch pState.State {
		case DeviceIsDisconnected:
			pss.discoverPeripheral(pName, pss.setDeviceConnected, pss.setDeviceDisconnected)
		case DeviceIsDiscovering:
			// We need to set a device into discovery mode so we won't retrigger discoveries
			// and pool them up
		case DeviceIsConnected:
			// Need to check if its streaming, if its not streaming than we have to do a healthcheck
			pss.Logger.Debug("Device is connected. Normal", "device", pName)
			pss.handleDeviceConnected(pName)
		case DeviceIsUnconfigured:
			pss.Logger.Debug("Device is still being configured.", "device", pName)
			pss.handleDeviceIsUnconfigured(pName)
		case DeviceIsConfiguring:
			// Do nothing configuration is happening in the background
		case DeviceIsConfigured:
			// Nothing to do here
			pss.handleDeviceIsConfigured(pName)
		case DeviceIsStreaming:
			//
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

		updatedState, err := pss.GetState(pName)
		if err != nil {
			// TODO: log something useful here
			continue
		}
		if updatedState.State == DeviceIsConnected ||
			updatedState.State == DeviceIsUnconfigured ||
			updatedState.State == DeviceIsConfigured {
			pss.Messages <- fmt.Sprintf(
				"Performing healthcheck. STATE: %s\n",
				DeviceStateToStr(updatedState.State),
			)
			pss.performHealthCheck(pName, updatedState)
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
