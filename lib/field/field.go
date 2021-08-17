package field

import (
	"crypto/rand"
	"encoding/binary"
	"io"
)

const (
	ModP  = uint32(2147483647) // 2^31 - 1
	Bytes = 4
)

func NegateVector(in []uint32) []uint32 {
	for i := range in {
		in[i] = ModP - in[i]
	}

	return in
}

func RandElementWithPRG(rnd io.Reader) uint32 {
	var buf [Bytes]byte
	_, err := rnd.Read(buf[:])
	if err != nil {
		panic("error in randomness")
	}
	return binary.BigEndian.Uint32(buf[:]) % ModP
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
		out[i] = binary.BigEndian.Uint32(buf[i*Bytes:(i+1)*Bytes]) % ModP
	}

	return out
}

func RandVector(length int) []uint32 {
	return RandVectorWithPRG(length, rand.Reader)
}

func ByteSliceToFieldElementSlice(in []byte) []uint32 {
	out := make([]uint32, len(in)/4)

	for i := range out {
		out[i] = binary.BigEndian.Uint32(in[i*Bytes:(i+1)*Bytes]) % ModP
	}

	return out
}
