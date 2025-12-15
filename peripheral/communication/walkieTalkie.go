package communication

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

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

// func (wt *WalkieTalkie) SendCommand(cmd Command, payload []byte) (int, error) {
// 	switch cmd {
// 	case CmdRequestID:
// 		return wt.RequestIdentification(d * serial.Port)
// 	}
// }

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

func (wt *WalkieTalkie) sendPacket(data any) error {
	// Prepare the payload
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, data)
	if err != nil {
		return err
	}

	// Send the payload
	_, err = wt.Serial.Write(buf.Bytes())
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

func (wt *WalkieTalkie) sendHeader(h *header) error {
	err := wt.sendPacket(h)
	if err != nil {
		return err
	}

	var ack packets.AckPacket
	err = wt.readPacket(&ack)
	if err != nil {
		return err
	}

	return nil
}

func (wt *WalkieTalkie) sendBody(payload any, resp packets.Packet) error {
	err := wt.sendPacket(payload)
	if err != nil {
		return err
	}

	err = wt.readPacket(resp)
	if err != nil {
		return err
	}

	return nil
}

func (wt *WalkieTalkie) SendCommand(cmd types.Command, payload any,
	responseBody packets.Packet) error {
	// Prepare the header
	header := header{
		StartByte: constvar.StartOfText,
		CMD:       cmd,
		EndByte:   constvar.EndOfText,
	}

	// Send the header
	err := wt.sendHeader(&header)
	if err != nil {
		return err
	}

	// Send the body
	err = wt.sendBody(payload, responseBody)
	if err != nil {
		return err
	}

	return nil
}
