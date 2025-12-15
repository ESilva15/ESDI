// Package packets will define the available packets we have
package packets

type Packet interface {
	Validate() bool
}

type PacketType uint8

const (
	identificationPacket PacketType = iota
)

const (
	identificationPacketStr = "identificationPacket"
	unknownPacketStr        = "unknownPacket"
)

func (p PacketType) String() string {
	switch p {
	case identificationPacket:
		return identificationPacketStr
	default:
		return unknownPacketStr
	}
}
