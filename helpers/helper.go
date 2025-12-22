// Package helper
package helper

import (
	"bytes"
	"encoding/binary"
)

func B32(s string) [32]byte {
	var b [32]byte
	copy(b[:], s)
	return b
}

func StructToBytes(s any) ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, s)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}
