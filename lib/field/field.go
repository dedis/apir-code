package field

import (
	"encoding/binary"
	"io"

	"github.com/si-co/vpir-code/lib/utils"
)

const (
	ModP                 = uint32(2147483647) // 2^31 - 1
	Bytes                = 4
	Bits                 = 31
	Mask                 = 127
	ConcurrentExecutions = 4
)

func NegateVector(in []uint32) []uint32 {
	for i := range in {
		in[i] = ModP - in[i]
	}

	return in
}

// Element converts input bytes into a field toElement.
// It sources a new random byte string in the rare case
// when the input bytes convert to ModP
func toElement(in []byte) uint32 {
	// Clearing the top most bit of uint32
	in[0] &= Mask
	// Take the first Bytes bytes
	//out := *(*uint32)(unsafe.Pointer(&in))
	out := binary.BigEndian.Uint32(in[:Bytes])
	// Make sure that toElement is not equal 2^31 - 1
	for out == ModP {
		var buf [Bytes]byte
		_, err := utils.RandomPRG().Read(buf[:])
		if err != nil {
			panic("error in randomness")
		}
		out = binary.BigEndian.Uint32(buf[:])
	}
	return out
}

func RandElementWithPRG(rnd io.Reader) uint32 {
	var buf [Bytes]byte
	var out = ModP
	// Make sure that toElement is not equal 2^31 - 1
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
	return RandElementWithPRG(utils.RandomPRG())
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
	return RandVectorWithPRG(length, utils.RandomPRG())
}

func BytesToElements(out []uint32, in []byte) {
	//if len(in) % Bytes != 0 {
	//	padding := make([]byte, Bytes-len(in) % Bytes)
	//	in = append(in, padding...)
	//}
	for i := range out {
		out[i] = toElement(in[i*Bytes : (i+1)*Bytes])
	}
}

// VectorToBytes extracts bytes from a vector of field elements.  Assume that
// only 3 bytes worth of data are embedded in each field toElement and therefore
// strips the initial zero from each field toElement.
func VectorToBytes(in interface{}) []byte {
	switch vec := in.(type) {
	case []uint32:
		elemSize := Bytes - 1
		out := make([]byte, len(vec)*elemSize)
		for i, e := range vec {
			fieldBytes := make([]byte, Bytes)
			binary.BigEndian.PutUint32(fieldBytes, e)
			// strip first zero and copy to the output
			copy(out[i*elemSize:(i+1)*elemSize], fieldBytes[1:])
		}
		return out
	default:
		return nil
	}
}
