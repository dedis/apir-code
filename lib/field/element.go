package field

import (
	"encoding/binary"
	"fmt"
	"io"
)

var p = uint32(2147483647) // 2^31 - 1

type Element struct {
	E uint32
}

const Bytes = 4

// Set z = x
func (z *Element) Set(x *Element) *Element {
	z.E = x.E
	return z
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	z.E = 0
	return z
}

// SetOne z = 1
func (z *Element) SetOne() *Element {
	z.E = 1
	return z
}

// Equal returns z == x
func (z *Element) Equal(x *Element) bool {
	return z.E == x.E
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	return z.E == 0
}

// SetRandom sets z to a random element < q
func (z *Element) SetRandom(rnd io.Reader) (*Element, error) {
	b := make([]byte, Bytes)
	if _, err := io.ReadFull(rnd, b); err != nil {
		return nil, err
	}

	z.E = binary.LittleEndian.Uint32(b)

	// mod if necessary
	if z.E >= p {
		z.E -= p
	}

	return z, nil
}

// RandomVector returns a vector composed of length random field elements
func RandomVector(rnd io.Reader, length int) ([]Element, error) {
	bytesLength := length*Bytes + 1
	bytes := make([]byte, bytesLength)
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}

	zs := make([]Element, length)
	for i := 0; i < length; i++ {
		buf := make([]byte, Bytes)
		copy(buf, bytes[i*Bytes:(1+i)*Bytes])
		zs[i].SetBytes(buf)
	}

	return zs, nil
}

// RandomVectorPointers returns a vector composed of length random pointers
// to field elements
func RandomVectorPointers(rnd io.Reader, length int) ([]*Element, error) {
	var e Element
	bytesLength := length*16 + 1
	bytes := make([]byte, bytesLength)
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}
	zs := make([]*Element, length)
	for i := 0; i < length; i++ {
		e.SetBytes(bytes[i*Bytes : (i+1)*Bytes])
		zs[i] = &e
	}

	return zs, nil
}

// ZeroVector returns a vector of zero elements
func ZeroVector(length int) []Element {
	zeroVector := make([]Element, length)
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
			b := make([]byte, Bytes)
			binary.LittleEndian.PutUint32(b, e.E)
			copy(out[i*elemSize:(i+1)*elemSize], b[1:])
		}
		return out
	default:
		return nil
	}
}

// One returns 1
func One() Element {
	return Element{E: 1}
}

// Zero returns 0
func Zero() Element {
	return Element{E: 0}
}

// Mul z = x * y mod q
func (z *Element) Mul(x, y *Element) *Element {
	//(ab >> 31) + ab & (mask 31 bits) mod P
	zz := (uint64(x.E) * uint64(y.E)) % uint64(p)
	z.E = uint32(zz)
	return z
}

// Add z = x + y mod q
func (z *Element) Add(x, y *Element) *Element {
	z.E = (x.E + y.E) % p
	//if z.E >= p {
	//z.E -= p
	//}

	return z
}

// Sub  z = x - y mod q
func (z *Element) Sub(x, y *Element) *Element {
	z.E = x.E - y.E
	if z.E < 0 {
		z.E += p
	}

	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	z.E = p - x.E
	return z
}

// String returns the string form of an Element in Montgomery form
func (z *Element) String() string {
	return fmt.Sprint(z.E)
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value (in Montgomery form), and returns z.
func (z *Element) SetBytes(in []byte) *Element {
	if len(in) == Bytes {
		z.E = binary.LittleEndian.Uint32(in)
		if z.E >= p {
			z.E -= p
		}

		return z
	}

	z.E = binary.LittleEndian.Uint32(in[:Bytes])

	if z.E >= p {
		z.E -= p
	}

	return z
}

// TODO: change this, here to 16 for compatibility reasons
func (z *Element) SetFixedLengthBytes(e [16]byte) *Element {
	z.SetBytes(e[:])
	return z
}
