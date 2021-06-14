package field

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
)

type Element struct {
	E uint64
}

const Bytes = 8

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
	//z.E = 1 << 6
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
	z.E = binary.LittleEndian.Uint64(b)
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
		buf := make([]byte, 8)
		copy(buf, bytes[i*Bytes:(1+i)*Bytes])
		zs[i].SetBytes(buf)
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
			b := make([]byte, 4)
			binary.LittleEndian.PutUint64(b, e.E)
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
func (z *Element) Mul(a, b *Element) *Element {
	if bits.OnesCount64(a.E) == 0 || bits.OnesCount64(b.E) == 0 {
		z.E = 0
		return z
	}

	first := a.E
	second := b.E
	result := uint64(0)
	for i := 0; i < 64; i++ {
		ab0 := first &^ (second&1 - 1)
		ra63 := (1<<4 | 1<<3 | 1<<1 | 1<<0) &^ (first>>(64-1) - 1)

		result, first, second = result^ab0, first<<1^ra63, second>>1
	}

	z.E = result
	return z
}

// Add z = x + y mod q
func (z *Element) Add(x, y *Element) *Element {
	z.E = x.E ^ y.E
	return z
}

// Sub  z = x - y mod q
func (z *Element) Sub(x, y *Element) *Element {
	z.E = x.E ^ y.E
	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	return x
}

// String returns the string form of an Element in Montgomery form
func (z *Element) String() string {
	return fmt.Sprint(z.E)
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value (in Montgomery form), and returns z.
func (z *Element) SetBytes(in []byte) *Element {
	if len(in) == 8 {
		z.E = binary.LittleEndian.Uint64(in)
		return z
	}

	//sum := blake2b.Sum256(in)
	z.E = binary.LittleEndian.Uint64(in[:8])
	return z
}

// TODO: change this, here to 16 for compatibility reasons
func (z *Element) SetFixedLengthBytes(e [16]byte) *Element {
	z.SetBytes(e[:])
	return z
}
