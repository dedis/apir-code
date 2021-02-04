// Package field contains field arithmetic operations for modulus 2^32
package field

import (
	"encoding/binary"
	"io"
	"math/bits"
	"strconv"
)

type Element uint32

// Bits number bits needed to represent Element
const Bits = 32

// Bytes number bytes needed to represent Element
const Bytes = 4

// Set z = x
func (z *Element) Set(x *Element) *Element {
	*z = *x
	return z
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	*z = 0
	return z
}

// SetOne z = 1
func (z *Element) SetOne() *Element {
	*z = 1
	return z
}

// Equal returns z == x
func (z *Element) Equal(x *Element) bool {
	return (*z == *x)
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	return *z == 0
}

// SetRandom sets z to a random element < q
func (z *Element) SetRandom(rnd io.Reader) (*Element, error) {
	var bytes [4]byte
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}
	*z = Element(binary.BigEndian.Uint32(bytes[:]))
	return z, nil
}

// PowerVectorWithOne returns vector (1, alpha, ..., alpha^(length))
func PowerVectorWithOne(alpha Element, length int) []Element {
	a := make([]Element, length+1)
	a[0] = One()
	a[1] = alpha
	for i := 2; i < len(a); i++ {
		a[i] = (a[i-1] * alpha) % (0xFFFFFFFF - 1)
	}

	return a
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
		var buf [4]byte
		copy(buf[:], bytes[i*Bytes:(1+i)*Bytes])
		zs[i] = Element(binary.BigEndian.Uint32(buf[:]))

	}

	return zs, nil
}

func RandomVectors(rnd io.Reader, vectorLen, blockLen int) ([][]Element, error) {
	bytesLength := (vectorLen*blockLen + 1) * Bytes
	bytes := make([]byte, bytesLength)
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}
	zs := make([][]Element, vectorLen)
	pos := 0
	for i := 0; i < vectorLen; i++ {
		zs[i] = make([]Element, blockLen)
		for j := 0; j < blockLen; j++ {
			zs[i][j] = Element(binary.BigEndian.Uint32(bytes[pos : pos+Bytes]))
			pos += Bytes
		}
	}
	return zs, nil
}

// ZeroVector returns a vector of zero elements
func ZeroVector(length int) []Element {
	// default value is already zero
	out := make([]Element, length)
	return out
}

// VectorToBytes extracts bytes from a vector of field elements.
func VectorToBytes(in []Element) []byte {
	out := make([]byte, 0, len(in)*(Bytes-1))
	for _, e := range in {
		fieldBytes := e.Bytes()
		// strip first zero
		out = append(out, fieldBytes[:]...)
	}

	return out
}

// One returns 1 (in montgommery form)
func One() Element {
	return Element(uint32(1))
}

// Zero returns 0
func Zero() Element {
	return Element(uint32(0))
}

// TODO
// Mul z = x * y mod q
func (z *Element) Mul(x, y *Element) *Element {
	//*z = (*x * *y) % 4294967295
	*z = (*x * *y)
	return z
}

// Add z = x + y mod q
func (z *Element) Add(x, y *Element) *Element {
	out, _ := bits.Add32(uint32(*x), uint32(*y), 0)
	*z = Element(out)
	return z
}

// Add returns x + y mod q
func Add(x, y Element) Element {
	out, _ := bits.Add32(uint32(x), uint32(y), 0)
	return Element(out)
}

// Sub  z = x - y mod q
func (z *Element) Sub(x, y *Element) *Element {
	diff, _ := bits.Sub32(uint32(*x), uint32(*y), 0)
	*z = Element(diff)
	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	diff, _ := bits.Sub32(uint32(0), uint32(*x), 0)
	*z = Element(diff)
	return z
}

// String returns the string form of an Element in Montgomery form
func (z *Element) String() string {
	return strconv.FormatUint(uint64(*z), 10)
}

func (z *Element) HexString() string {
	return strconv.FormatUint(uint64(*z), 16)
}

// Bytes returns the value
// of z as a big-endian byte array.
func (z *Element) Bytes() [4]byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(*z))
	return b
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value and returns z.
func (z *Element) SetBytes(e []byte) *Element {
	out := binary.BigEndian.Uint32(e)
	*z = Element(out)
	return z
}

func (z *Element) SetFixedLengthBytes(e [16]byte) *Element {
	out := binary.BigEndian.Uint32(e[:])
	*z = Element(out)
	return z
}
