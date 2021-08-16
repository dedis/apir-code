package field

import (
	"crypto/rand"
	"encoding/binary"
	"io"
)

const ModP = uint32(2147483647) // 2^31 - 1

func RandElementWithPRG(rnd io.Reader) uint32 {
	var buf [4]byte
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
	bytesLength := length * 4
	buf := make([]byte, bytesLength)
	_, err := rnd.Read(buf)
	if err != nil {
		panic("error in randomness")
	}
	out := make([]uint32, length)
	for i := range out {
		out[i] = binary.BigEndian.Uint32(buf[i*4:(i+1)*4]) % ModP
	}

	return out
}

func RandVector(length int) []uint32 {
	return RandVectorWithPRG(length, rand.Reader)
}

func NegateVector(in []uint32) []uint32 {
	for i := range in {
		in[i] = ModP - in[i]
	}

	return in
}
