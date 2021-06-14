package field

import (
	"encoding/binary"
	"io"

	"lukechampine.com/uint128"
)

type Element struct {
	E uint128.Uint128
}

const Bytes = 16

func New(e uint128.Uint128) Element {
	return Element{E: e}
}

// Set z = x
func (z *Element) Set(x *Element) *Element {
	z.E.Lo = x.E.Lo
	z.E.Hi = x.E.Hi
	return z
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	z.E.Lo = 0
	z.E.Hi = 0
	return z
}

// SetOne z = 1
func (z *Element) SetOne() *Element {
	z.E.Lo = 1
	z.E.Hi = 0
	return z
}

// Equal returns z == x
func (z *Element) Equal(x *Element) bool {
	return z.E.Equals(x.E)
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	return z.E.IsZero()
}

// SetRandom sets z to a random element < q
func (z *Element) SetRandom(rnd io.Reader) (*Element, error) {
	var bytes [16]byte
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}
	z.SetBytes(bytes[:])
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
			fieldBytes := [16]byte{}
			binary.LittleEndian.PutUint64(fieldBytes[:8], e.E.Lo)
			binary.LittleEndian.PutUint64(fieldBytes[8:], e.E.Hi)
			// strip first zero and copy to the output
			//copy(out[i*elemSize:(i+1)*elemSize], fieldBytes[1:])
			// TODO
			copy(out[i*elemSize:(i+1)*elemSize], fieldBytes[:])
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
	z.E = x.E.MulWrap(y.E)
	return z
}

// Add z = x + y mod q
func (z *Element) Add(x, y *Element) *Element {
	z.E = x.E.AddWrap(y.E)
	return z
}

// Sub  z = x - y mod q
func (z *Element) Sub(x, y *Element) *Element {
	z.E = x.E.SubWrap(y.E)
	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	z.E = uint128.Max.SubWrap(x.E)
	return z
}

// String returns the string form of an Element in Montgomery form
func (z *Element) String() string {
	return z.E.String()
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value (in Montgomery form), and returns z.
func (z *Element) SetBytes(in []byte) *Element {
	el := new(Element)
	el.E = uint128.New(0, 0)
	el.E.PutBytes(in)
	return el
}

func (z *Element) SetFixedLengthBytes(e [16]byte) *Element {
	z.SetBytes(e[:])
	return z
}
