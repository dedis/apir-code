package utils

import (
	"encoding/binary"
	"math/bits"
)

func Uint32SliceToByteSlice(in []uint32) []byte {
	nb := 4
	out := make([]byte, len(in)*nb)
	for i := range in {
		binary.BigEndian.PutUint32(out[i*nb:(i+1)*nb], in[i])
	}

	return out
}

func ByteSliceToUint32Slice(in []byte) []uint32 {
	nb := 4
	out := make([]uint32, len(in)/nb)
	for i := range out {
		out[i] = binary.BigEndian.Uint32(in[i*nb:])
	}
	return out
}

func ByteToBits(data []byte) []bool {
	out := make([]bool, len(data)*8)
	for i, d := range data {
		for j := 0; j < 8; j++ {
			if bits.LeadingZeros8(d) == 0 {
				// No leading 0 means that it is a 1
				out[i*8+j] = true
			} else {
				out[i*8+j] = false
			}
			d = d << 1
		}
	}
	return out
}
