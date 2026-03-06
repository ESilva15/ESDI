package telemetry

import (
	"encoding/binary"
	"testing"
)

var (
	testRaw uint64 = 0x1122334455667788
	sink    []byte
)

// The "Manual" way (with the fix for the missing 24-bit shift)
func BenchmarkPackManual(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 8)
		buf = append(buf,
			uint8(testRaw),
			uint8(testRaw>>8),
			uint8(testRaw>>16),
			uint8(testRaw>>24),
			uint8(testRaw>>32),
			uint8(testRaw>>40),
			uint8(testRaw>>48),
			uint8(testRaw>>56),
		)
		sink = buf
	}
}

// The "Loop" way
func BenchmarkPackLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 8)
		for j := 0; j < 8; j++ {
			buf = append(buf, uint8(testRaw>>(j*8)))
		}
		sink = buf
	}
}

// The "Binary + Stack Array" way (Best Practice)
func BenchmarkPackBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 8)
		var temp [8]byte
		binary.LittleEndian.PutUint64(temp[:], testRaw)
		buf = append(buf, temp[:]...)
		sink = buf
	}
}
