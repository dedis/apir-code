package field

import (
	"encoding/binary"
	"io"
	"unsafe"

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

// TODO: FIX THIS, NOT FIELD ELEMENT
func ByteSliceToFieldElementSlice(out []uint32, in []byte) {
	// for i := range out {
	// 	out[i] = binary.BigEndian.Uint32(in[i*Bytes:(i+1)*Bytes]) % ModP
	// }
	out = *(*[]uint32)(unsafe.Pointer(&in))
}

// VectorToBytes extracts bytes from a vector of field elements.  Assume that
// only 3 bytes worth of data are embedded in each field element and therefore
// strips the initial zero from each field element.
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
