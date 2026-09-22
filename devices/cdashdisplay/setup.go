package cdashdisplay

import (
	"fmt"

	"esdi/constants"
)

func (cds *CDashDisplay) setupForIracing() error {
	err := cds.LoadLayout("layout.yaml")
	if err != nil {
		return err
	}

	return nil
}

func (cds *CDashDisplay) setupForBeamNG() error {
	err := cds.LoadLayout("beamng.yaml")
	if err != nil {
		return err
	}

	return nil
}

func (cds *CDashDisplay) Setup(provider string) error {
	switch provider {
	case constants.IRacingProviderName:
		return cds.setupForIracing()
	case constants.BeamNGProviderName:
		return cds.setupForBeamNG()
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}
}
