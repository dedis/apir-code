package utils

import (
	"encoding/binary"
	"math/bits"
)

func Uint32SliceToByteSlice(in []uint32) []byte {
	out := make([]byte, len(in)*4)
	for i := range in {
		binary.BigEndian.PutUint32(out[i*4:(i+1)*4], in[i])
	}

	return out
}

func ByteSliceToUint32Slice(in []byte) []uint32 {
	out := make([]uint32, len(in)/4)
	for i := range out {
		out[i] = binary.BigEndian.Uint32(in[i*4:])
	}
	return out
}

func Uint64SliceToByteSlice(in []uint64) []byte {
	nb := 8
	out := make([]byte, len(in)*nb)
	for i := range in {
		binary.BigEndian.PutUint64(out[i*nb:(i+1)*nb], in[i])
	}

	return out
}

func ByteSliceToUint64Slice(in []byte) []uint64 {
	nb := 8
	out := make([]uint64, len(in)/nb)
	for i := range out {
		out[i] = binary.BigEndian.Uint64(in[i*nb:])
	}
	return out
}

func ByteToBits(data []byte) []bool {
	out := make([]bool, len(data)*8) // Performance x 2 as no append occurs.
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
