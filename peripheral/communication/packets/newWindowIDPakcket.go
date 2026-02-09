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
