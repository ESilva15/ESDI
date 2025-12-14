package communication

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

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

		if b[0] == StartOfText {
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

func (wt *WalkieTalkie) SendCommand(payload []byte) (int, error) {
	return wt.Serial.Write(payload)
}
