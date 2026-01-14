package packets

import (
	"esdi/peripheral/communication/constvar"
)

type IdentificationPacket struct {
	StartMarker byte
	DeviceID    uint8
	PktType     PacketType
	Name        [32]byte
	EndMarker   byte
}

func (pkt *IdentificationPacket) Validate() bool {
	// fmt.Println("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
	// fmt.Println(pkt.StartMarker)
	// fmt.Println(pkt.PktType.String())
	// fmt.Println(string(pkt.Name[:]))
	// fmt.Println(pkt.EndMarker)
	// fmt.Println("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")

	// Change this to return an error and add a validation against available
	// device ids

	if pkt.StartMarker != constvar.StartOfText ||
		pkt.EndMarker != constvar.EndOfText {
		return false
	}

	return true
}
