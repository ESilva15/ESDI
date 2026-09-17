package cdashdisplay

import (
	"fmt"
	"log/slog"
	"time"

	"esdi/peripheral/communication"
	"esdi/peripheral/communication/packets"

	"github.com/tarm/serial"
	portp "go.bug.st/serial"
)

func listPorts() ([]string, error) {
	ports, err := portp.GetPortsList()
	if err != nil {
		return []string{}, err
	}

	return ports, nil
}

func probe(WT *communication.WalkieTalkie) error {
	err := WT.TurnOn()
	if err != nil {
		return err
	}

	// Send the identification command
	cmd := communication.CmdRequestID
	var response packets.IdentificationPacket
	err = WT.SendCommand(cmd, []byte{0x06, 0x07, 0x08, 0x09}, &response)
	if err != nil {
		return err
	}

	if response.DeviceID != 0x01 {
		return fmt.Errorf("wrong ID")
	}

	return nil
}

func findDisplayPort() (*communication.WalkieTalkie, error) {
	ports, err := listPorts()
	if err != nil {
		return nil, err
	}

	slog.Info(fmt.Sprintf("Looking into %v", ports))

	var wt *communication.WalkieTalkie
	for _, port := range ports {
		slog.Info(fmt.Sprintf("Trying port %s", port))

		wt = &communication.WalkieTalkie{
			Cfg: &serial.Config{
				Name:        port,
				Baud:        115200,
				ReadTimeout: 500 * time.Millisecond,
			},
		}

		slog.Info(fmt.Sprintf("Started probing port %s", port))

		probeResult := make(chan error, 1)

		go func() {
			probeResult <- probe(wt)
		}()

		select {
		case err = <-probeResult:
			// Probe completed normally (could be success or error)
		case <-time.After(2 * time.Second):
			// Hard timeout reached
			err = fmt.Errorf("probe completely hung/timed out: %s", port)
		}

		slog.Info(fmt.Sprintf("Finished probing port %s", port))

		if err == nil {
			slog.Info(fmt.Sprintf("Success probing port %s: %+v", port, err))
			break
		}

		slog.Info(fmt.Sprintf("wasn't port %s", port))
		wt = nil
	}

	if wt == nil {
		return nil, fmt.Errorf("couldn't find cdashdisplay")
	}

	slog.Info(fmt.Sprintf("found cdashdisplay on port: %s", wt.Cfg.Name))
	return wt, nil
}
