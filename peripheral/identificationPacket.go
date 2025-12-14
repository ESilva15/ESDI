package peripheral

import (
	"encoding/binary"
	comm "esdi/peripheral/communication"
	"fmt"
)

type IdentificationPacket struct {
	StartMarker byte
	DeviceID    uint8
	PktType     PacketType
	Name        [32]byte
	EndMarker   byte
}

func (pkt *IdentificationPacket) Read(wt *comm.WalkieTalkie) error {
	err := wt.ReadFramedData(binary.Size(IdentificationPacket{}), pkt)
	if err != nil {
		return err
	}

	if !pkt.Validate() {
		return fmt.Errorf("invalid data")
	}

	return nil
}

func (pkt *IdentificationPacket) Validate() bool {
	fmt.Println("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
	fmt.Println(pkt.StartMarker)
	fmt.Println(pkt.PktType.String())
	fmt.Println(string(pkt.Name[:]))
	fmt.Println(pkt.EndMarker)
	fmt.Println("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")

	// Change this to return an error and add a validation against available
	// device ids

	if pkt.StartMarker != comm.StartOfText ||
		pkt.EndMarker != comm.EndOfText {
		return false
	}

	return true
}
