package peripheral

import (
	"encoding/binary"
	"fmt"
)

type IdentificationPacket struct {
	StartMarker byte
	PktType     PacketType
	Name        [32]byte
	EndMarker   byte
}

func (pkt *IdentificationPacket) Read(wt *WalkieTalkie) error {
	fmt.Println("Reading framed data")

	err := wt.ReadFramedData(binary.Size(IdentificationPacket{}), pkt)
	if err != nil {
		return err
	}

	if !pkt.Validate() {
		return fmt.Errorf("invalid packet")
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

	if pkt.StartMarker != startOfText ||
		pkt.EndMarker != endOfText {
		return false
	}

	return true
}
