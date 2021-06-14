package field

/*
#cgo CFLAGS: -fopenmp -O2
#cgo LDFLAGS: -lcrypto -lm -fopenmp
#include "../c/dpf.h"
#include "../c/dpf.c"
*/
import "C"
import (
	"bytes"
	"io"
	"unsafe"
)

type Element [16]byte

const Limbs = 2

const Bytes = 16

// Set z = x
func (z *Element) Set(x *Element) *Element {
	*z = *x
	return z
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	// already initialized to zero
	zero := [16]byte{}
	*z = zero
	return z
}

// SetOne z = 1
func (z *Element) SetOne() *Element {
	z.SetZero()
	(*z)[0] = 1
	return z
}

// Equal returns z == x
func (z *Element) Equal(x *Element) bool {
	return bytes.Equal((*z)[:], (*x)[:])
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	zero := make([]byte, 16)
	return bytes.Equal((*z)[:], zero)
}

// SetRandom sets z to a random element < q
func (z *Element) SetRandom(rnd io.Reader) (*Element, error) {
	var bytes [16]byte

	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}

	zero := Zero()
	return z.Add(z, &zero), nil
}

func RandomVector(rnd io.Reader, length int) ([]Element, error) {
	bytesLength := length*Bytes + 1
	bytes := make([]byte, bytesLength)
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}

	zs := make([]Element, length)
	for i := 0; i < length; i++ {
		var buf [16]byte
		copy(buf[:], bytes[i*Bytes:(1+i)*Bytes])
		zs[i].SetFixedLengthBytes(buf)
	}

	return zs, nil
}

// ZeroVector returns a vector of zero elements
func ZeroVector(length int) []Element {
	zeroVector := make([]Element, length)
	for i := range zeroVector {
		zero := Zero()
		zeroVector[i] = zero
	}
	return zeroVector
}

// VectorToBytes extracts bytes from a vector of field elements.  Assume that
// only 15 bytes worth of data are embedded in each field element and therefore
// strips the initial zero from each byte.
func VectorToBytes(in interface{}) []byte {
	switch vec := in.(type) {
	case []Element:
		elemSize := Bytes - 1
		out := make([]byte, len(vec)*elemSize)
		for i, e := range vec {
			fieldBytes := e.Bytes()
			// strip first zero and copy to the output
			copy(out[i*elemSize:(i+1)*elemSize], fieldBytes[1:])
		}
		return out
	default:
		return nil
	}
}

// One returns 1
func One() Element {
	var one Element
	one.SetOne()
	return one
}

// Zero returns 0
func Zero() Element {
	var zero Element
	zero.SetZero()
	return zero
}

// Mul z = x * y mod q
func (z *Element) Mul(x, y *Element) *Element {
	res := C.multModP([16]byte(*x), [16]byte(*y))
	out := C.GoBytes(unsafe.Pointer(&res), 16)
	copy((*z)[:], out)
	return z
}

// Add z = x + y mod q
func (z *Element) Add(x, y *Element) *Element {
	res := C.addModP([16]byte(*x), [16]byte(*y))
	out := C.GoBytes(unsafe.Pointer(&res), 16)
	copy((*z)[:], out)
	return z
}

// Sub  z = x - y mod q
func (z *Element) Sub(x, y *Element) *Element {
	res := C.subModP([16]byte(*x), [16]byte(*y))
	out := C.GoBytes(unsafe.Pointer(&res), 16)
	copy((*z)[:], out)
	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	res := C.negModP([16]byte(*x))
	out := C.GoBytes(unsafe.Pointer(&res), 16)
	copy((*z)[:], out)
	return z
}

// Bytes returns the value of z as a big-endian byte array.
func (z *Element) Bytes() (res [Limbs * 8]byte) {
	return [16]byte(*z)
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value, and returns z.
func (z *Element) SetBytes(e []byte) *Element {
	bytes := [16]byte{}
	copy(bytes[:], e)
	return z.SetFixedLengthBytes(bytes)
}

func (z *Element) SetFixedLengthBytes(e [16]byte) *Element {
	*z = e
	return z
}
