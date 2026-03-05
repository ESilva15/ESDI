// Package helper
package helper

import (
	"bytes"
	"encoding/binary"
)

type Vector struct {
	DX uint16
	DY uint16
}

// type MultiError struct {
// 	Errors []error
// }

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

//	func CopyBytes(dest []byte, destSize int, src string) {
//		copy(dest[:], []byte(src))
//		dest[min(destSize-1, len(src))] = '\x00'
//	}

func CopyBytes(dest []byte, src string) {
	// 1. Clear the destination (optional but safer for fixed-width telemetry)
	for i := range dest {
		dest[i] = 0
	}

	// 2. Copy as much as fits, leaving at least 1 byte for a null terminator
	// We limit the copy to len(dest) - 1
	n := copy(dest[:len(dest)-1], src)

	// 3. Explicitly null terminate after the last written byte
	dest[n] = '\x00'
}
