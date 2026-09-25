package packets

import (
	"esdi/peripheral/communication/constvar"
)

type NewWindowID struct {
	StartMarker byte
	ID          int16
	EndMarker   byte
}

func (pkt *NewWindowID) Validate() bool {
	if pkt.StartMarker != constvar.StartOfText ||
		pkt.EndMarker != constvar.EndOfText {
		return false
	}

	return true
}

type HealthCheck struct {
	StartMarker byte
	Response    byte
	EndMarker   byte
}

func (pkt *HealthCheck) Validate() bool {
	if pkt.StartMarker != constvar.StartOfText ||
		pkt.EndMarker != constvar.EndOfText {
		return false
	}

	if pkt.Response != 0x06 {
		return false
	}

	return true
}
