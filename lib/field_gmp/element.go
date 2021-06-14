package field

import (
	"io"
)

type Element struct {
	e *Int
}

const Bytes = 16

var _modulus Int

func init() {
	m := NewInt(0)
	_modulus = *m
	err := (&_modulus).SetString("170141183460469231731687303715884105727", 10)
	if err != nil {
		panic(err)
	}
}

func New(b []byte) Element {
	el := Element{e: NewInt(0)}
	el.e.SetBytes(b)
	return el
}

// Set z = x
func (z *Element) Set(x *Element) *Element {
	if z.e == nil {
		z.e = NewInt(0)
	}
	z.e.Set(x.e)
	return z
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	if z.e == nil {
		z.e = NewInt(0)
	}
	z.e.SetInt64(0)
	return z
}

// SetOne z = 1
func (z *Element) SetOne() *Element {
	if z.e == nil {
		z.e = NewInt(0)
	}
	z.e.SetInt64(1)
	return z
}

// Equal returns z == x
func (z *Element) Equal(x *Element) bool {
	return CmpInt(z.e, x.e) == 0
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	zero := Zero()
	return z.Equal(&zero)
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
		zs[i] = One()
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
			fieldBytes := e.e.Bytes()
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
	one.e = NewInt(1)
	return one
}

// Zero returns 0
func Zero() Element {
	var zero Element
	zero.e = NewInt(0)
	return zero
}

// Mul z = x * y mod q
func (z *Element) Mul(x, y *Element) *Element {
	if z.e == nil {
		z.e = NewInt(1)
	}
	z.e.Mul(x.e, y.e)
	z.e.Mod(z.e, &_modulus)
	return z
}

// Add z = x + y mod q
func (z *Element) Add(x, y *Element) *Element {
	if z.e == nil {
		z.e = NewInt(0)
	}
	z.e.Add(x.e, y.e)
	z.e.Mod(z.e, &_modulus)
	return z
}

// Sub  z = x - y mod q
func (z *Element) Sub(x, y *Element) *Element {
	if z.e == nil {
		z.e = NewInt(0)
	}
	z.e.Sub(x.e, y.e)
	z.e.Mod(z.e, &_modulus)
	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	z.e.Sub(&_modulus, x.e)
	z.e.Mod(z.e, &_modulus)
	return z
}

// String returns the string form of an Element in Montgomery form
func (z *Element) String() string {
	return z.e.String()
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value (in Montgomery form), and returns z.
func (z *Element) SetBytes(in []byte) *Element {
	if z.e == nil {
		z.e = NewInt(0)
	}
	z.e.SetBytes(in)
	return z
}

func (z *Element) SetFixedLengthBytes(e [16]byte) *Element {
	z.SetBytes(e[:])
	return z
}
