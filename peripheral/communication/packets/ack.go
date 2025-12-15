package packets

import "esdi/peripheral/communication/constvar"

type AckPacket struct {
	StartMarker byte
	AckByte     byte
	EndMarker   byte
}

func (pkt *AckPacket) Validate() bool {
	if pkt.StartMarker != constvar.StartOfText ||
		pkt.EndMarker != constvar.EndOfText {
		return false
	}

	if pkt.AckByte != constvar.ACK {
		return false
	}

	return true
}
