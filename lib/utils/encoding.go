package utils

import "encoding/binary"

func ByteSliceToUint32Slice(in []byte) []uint32 {
	// since we suppose that this function is used only for decoding, we don't
	// check for non-integer division here
	out := make([]uint32, len(in)/4)
	for i := range out {
		out[i] = binary.BigEndian.Uint32(in[i*4 : (i+1)*4])
	}

	return out
}

func Uint32SliceToByteSlice(in []uint32) []byte {
	out := make([]byte, len(in)*4)
	for i := range in {
		binary.BigEndian.PutUint32(out[i*4:(i+1)*4], in[i])
	}

	return out
}
