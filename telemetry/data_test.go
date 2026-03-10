package telemetry

import (
	"bytes"
	"math"
	"testing"
)

func Test_TelemetryField(t *testing.T) {
	type TFTest struct {
		name   string
		tf     TelemetryField
		expect []byte
	}

	tests := []TFTest{
		{
			name: "test_max_uint8",
			tf: TelemetryField{
				Type: DataTypeUINT8,
				Raw:  uint64(math.MaxUint8),
			},
			expect: []byte{0x00, 0xFF},
		},
		{
			name: "test_max_int8",
			tf: TelemetryField{
				Type: DataTypeINT8,
				Raw:  uint64(math.MaxInt8),
			},
			expect: []byte{0x01, 0x7F},
		},
		{
			name: "test_max_uint16",
			tf: TelemetryField{
				Type: DataTypeUINT16,
				Raw:  uint64(math.MaxUint16),
			},
			expect: []byte{0x02, 0xFF, 0xFF},
		},
		{
			name: "test_max_int16",
			tf: TelemetryField{
				Type: DataTypeINT16,
				Raw:  uint64(math.MaxInt16),
			},
			expect: []byte{0x03, 0xFF, 0x7F},
		},
		{
			name: "test_string",
			tf: TelemetryField{
				Type: DataTypeSTRING,
				Str:  "a cool string!",
			},
			expect: []byte{0x08, 0x0E, 0x61, 0x20, 0x63, 0x6F, 0x6F, 0x6C, 0x20, 0x73,
				0x74, 0x72, 0x69, 0x6E, 0x67, 0x21},
		},
		{
			name: "test_char",
			tf: TelemetryField{
				Type: DataTypeCHAR,
				Raw:  uint64('R'),
			},
			expect: []byte{0x09, 0x52},
		},
	}

	bufPtr := bufferPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]

	for _, test := range tests {
		result := test.tf.Pack(buf)

		if !bytes.Equal(result, test.expect) {
			t.Errorf("\nTest: %s\nExpected: %v\nGot: %v\n", test.name, test.expect, result)
		}
	}
}

func Test_TelemetryData(t *testing.T) {
	data := TelemetryData{
		ActiveBinds: []BoundField{
			{
				Key: "Speed",
				ID:  Speed,
			},
			{
				Key: "Gear",
				ID:  Gear,
			},
		},
	}

	data.Values[Speed] = TelemetryField{
		Type: DataTypeUINT16,
		Raw:  uint64(254),
	}

	data.Values[Gear] = TelemetryField{
		Type: DataTypeCHAR,
		Raw:  uint64('R'),
	}

	expect := []byte{0x02, 0xFE, 0x00, 0x09, 0x52}

	result := data.Pack()
	if !bytes.Equal(result, expect) {
		t.Errorf("\nExpected: %v\nGot: %v\n", expect, result)
	}
}
