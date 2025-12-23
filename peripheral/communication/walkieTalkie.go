package communication

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	helper "esdi/helpers"
	"esdi/peripheral/communication/constvar"
	"esdi/peripheral/communication/packets"
	"esdi/peripheral/types"

	"github.com/tarm/serial"
)

type WalkieTalkie struct {
	Serial *serial.Port
	Cfg    *serial.Config
}

func (wt *WalkieTalkie) ReadFramedData(size int, packet any) error {
	buf := make([]byte, size)

	for {
		b := make([]byte, 1)
		_, err := wt.Serial.Read(b)
		if err != nil {
			fmt.Println("dev read:", err.Error())
			return err
		}

		if b[0] == constvar.StartOfText {
			buf[0] = b[0]
			break
		}
	}

	_, err := io.ReadFull(wt.Serial, buf[1:])
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buf)
	err = binary.Read(reader, binary.LittleEndian, packet)
	if err != nil {
		return err
	}

	return nil
}

func (wt *WalkieTalkie) TurnOn() error {
	var err error

	wt.Serial, err = serial.OpenPort(wt.Cfg)
	if err != nil {
		return err
	}

	return nil
}

func (wt *WalkieTalkie) TurnOff() error {
	return wt.Serial.Close()
}

func (wt *WalkieTalkie) RequestIdentification() (int, error) {
	return wt.Serial.Write([]byte{uint8(CmdRequestID)})
}

func (wt *WalkieTalkie) AknowledgeIdentification() (int, error) {
	return wt.Serial.Write([]byte{uint8(CmdAckID)})
}

type header struct {
	StartByte uint8
	CMD       types.Command
	EndByte   uint8
}

type CMDDataPacket struct {
	StartMarker uint8
	CMD         types.Command
	Len         uint16
	Payload     []byte
	CRC         uint8
	EndMarker   uint8
}

func (cmdp *CMDDataPacket) Serialize() []byte {
	buf := make([]byte, 0, 1+1+2+len(cmdp.Payload)+1+1)

	buf = append(buf, cmdp.StartMarker)
	buf = append(buf, byte(cmdp.CMD))
	buf = append(buf, byte(cmdp.Len), byte(cmdp.Len>>8))
	buf = append(buf, cmdp.Payload...)
	buf = append(buf, cmdp.CRC)
	buf = append(buf, cmdp.EndMarker)

	return buf
}

func (wt *WalkieTalkie) sendPacket(cmd types.Command, data any) error {
	// Prepare the payload
	payload, err := helper.StructToBytes(data)
	if err != nil {
		return err
	}

	packet := CMDDataPacket{
		StartMarker: constvar.StartOfText,
		CMD:         cmd,
		Len:         uint16(len(payload)),
		Payload:     payload,
		CRC:         CRC8(payload),
		EndMarker:   constvar.EndOfText,
	}

	// Send the payload
	serializedPacket := packet.Serialize()
	fmt.Println(serializedPacket)

	_, err = wt.Serial.Write(serializedPacket)
	if err != nil {
		return err
	}

	return nil
}

func (wt *WalkieTalkie) readPacket(resp packets.Packet) error {
	size := binary.Size(resp)
	if size < 0 {
		return fmt.Errorf("invalid packet size")
	}

	err := wt.ReadFramedData(size, resp)
	if err != nil {
		return err
	}

	if !resp.Validate() {
		return fmt.Errorf("badly formatted response")
	}

	return nil
}

// func (wt *WalkieTalkie) sendHeader(h *header) error {
// 	err := wt.sendPacket(h)
// 	if err != nil {
// 		return err
// 	}
//
// 	// var ack packets.AckPacket
// 	// err = wt.readPacket(&ack)
// 	// if err != nil {
// 	// 	return err
// 	// }
//
// 	return nil
// }

// func (wt *WalkieTalkie) sendBody(payload any, resp packets.Packet) error {
// 	err := wt.sendPacket(payload)
// 	if err != nil {
// 		return err
// 	}
//
// 	// err = wt.readPacket(resp)
// 	// if err != nil {
// 	// 	return err
// 	// }
//
// 	return nil
// }

func (wt *WalkieTalkie) SendCommand(cmd types.Command, payload any,
	responseBody packets.Packet,
) error {
	// Prepare the header
	// header := header{
	// 	StartByte: constvar.StartOfText,
	// 	CMD:       cmd,
	// 	EndByte:   constvar.EndOfText,
	// }

	// Send the header
	// err := wt.sendHeader(&header)
	// if err != nil {
	// 	return err
	// }

	// Send the body
	err := wt.sendPacket(cmd, payload)
	if err != nil {
		return err
	}

	// if the responseBody != nil, then we also read and populate it???
	if responseBody != nil {
		err = wt.readPacket(responseBody)
		if err != nil {
			return err
		}
	}

	return nil
}
