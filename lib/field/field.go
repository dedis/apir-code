package field

import (
	"crypto/rand"
	"encoding/binary"
	"io"
)

const (
	ModP  = uint32(2147483647) // 2^31 - 1
	Bytes = 4
	Bits  = 31
	Mask  = 127
)

func NegateVector(in []uint32) []uint32 {
	for i := range in {
		in[i] = ModP - in[i]
	}

	return in
}

func RandElementWithPRG(rnd io.Reader) uint32 {
	var buf [Bytes]byte
	var out = ModP
	// Make sure that element is not equal 2^31 - 1
	for out == ModP {
		_, err := rnd.Read(buf[:])
		if err != nil {
			panic("error in randomness")
		}
		// Clearing the top most bit of uint32
		buf[0] &= Mask
		out = binary.BigEndian.Uint32(buf[:])
	}
	return out
}

func RandElement() uint32 {
	return RandElementWithPRG(rand.Reader)
}

func RandVectorWithPRG(length int, rnd io.Reader) []uint32 {
	bytesLength := length * Bytes
	buf := make([]byte, bytesLength)
	_, err := rnd.Read(buf)
	if err != nil {
		panic("error in randomness")
	}
	out := make([]uint32, length)
	for i := range out {
		//Clearing the top most bit of uint32
		buf[i*Bytes] &= Mask
		out[i] = binary.BigEndian.Uint32(buf[i*Bytes : (i+1)*Bytes])
		for out[i] == ModP {
			out[i] = RandElementWithPRG(rnd)
		}
	}

	return out
}

func RandVector(length int) []uint32 {
	return RandVectorWithPRG(length, rand.Reader)
}

// TODO: fix this because of bias in randomness
func ByteSliceToFieldElementSlice(in []byte) []uint32 {
	out := make([]uint32, len(in)/4)

	for i := range out {
		out[i] = binary.BigEndian.Uint32(in[i*Bytes:(i+1)*Bytes]) % ModP
	}

	return out
}
