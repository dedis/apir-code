package utils

import "encoding/binary"

func Uint32SliceToByteSlice(in []uint32) []byte {
	out := make([]byte, len(in)*4)
	for i := range in {
		binary.BigEndian.PutUint32(out[i*4:(i+1)*4], in[i])
	}

	return out
}
