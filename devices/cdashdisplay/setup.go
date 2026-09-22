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
	// Doesn't matter which one we are picking, we need to reset the CDashDisplay
	// first - we will improve this setup behaviour later on with an Update to
	// change it without dropping connection or some shit
	err := cds.reset()
	if err != nil {
		return err
	}

	switch provider {
	case constants.IRacingProviderName:
		return cds.setupForIracing()
	case constants.BeamNGProviderName:
		return cds.setupForBeamNG()
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}
}
