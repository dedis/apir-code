// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by goff (v0.3.9) DO NOT EDIT

// Package field contains field arithmetic operations for modulus 170141183460469231731687303715884105727
package field

// /!\ WARNING /!\
// this code has not been audited and is provided as-is. In particular,
// there is no security guarantees such as constant time implementation
// or side-channel attack resistance
// /!\ WARNING /!\

import (
	"encoding/binary"
	"io"
	"math/big"
	"math/bits"
	"strconv"
	"sync"
)

// Element represents a field element stored on 2 words (uint64)
// Element are assumed to be in Montgomery form in all methods
// field modulus q =
//
// 170141183460469231731687303715884105727
type Element [2]uint64

// Limbs number of 64 bits words needed to represent Element
const Limbs = 2

// Bits number bits needed to represent Element
const Bits = 127

// Bytes number bytes needed to represent Element
const Bytes = Limbs * 8

// field modulus stored as big.Int
var _modulus big.Int

// Modulus returns q as a big.Int
// q =
//
// 170141183460469231731687303715884105727
func Modulus() *big.Int {
	return new(big.Int).Set(&_modulus)
}

// q (modulus)
var qElement = Element{
	18446744073709551615,
	9223372036854775807,
}

// rSquare
var rSquare = Element{
	4,
	0,
}

var bigIntDefault [Limbs]big.Word
var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int).SetBits(bigIntDefault[:])
	},
}

func init() {
	_modulus.SetString("170141183460469231731687303715884105727", 10)
	for i := 0; i < len(bigIntDefault); i++ {
		bigIntDefault[i] = big.Word(0x1)
	}

}

// SetUint64 z = v, sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
func (z *Element) SetUint64(v uint64) *Element {
	*z = Element{v}
	return z.Mul(z, &rSquare) // z.ToMont()
}

// Set z = x
func (z *Element) Set(x *Element) *Element {
	z[0] = x[0]
	z[1] = x[1]
	return z
}

// SetInterface converts i1 from uint64, int, string, or Element, big.Int into Element
// panic if provided type is not supported
func (z *Element) SetInterface(i1 interface{}) *Element {
	switch c1 := i1.(type) {
	case Element:
		return z.Set(&c1)
	case *Element:
		return z.Set(c1)
	case uint64:
		return z.SetUint64(c1)
	case int:
		return z.SetString(strconv.Itoa(c1))
	case string:
		return z.SetString(c1)
	case *big.Int:
		return z.SetBigInt(c1)
	case big.Int:
		return z.SetBigInt(&c1)
	case []byte:
		return z.SetBytes(c1)
	default:
		panic("invalid type")
	}
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	z[0] = 0
	z[1] = 0
	return z
}

// SetOne z = 1 (in Montgomery form)
func (z *Element) SetOne() *Element {
	z[0] = 2
	z[1] = 0
	return z
}

// Div z = x*y^-1 mod q
func (z *Element) Div(x, y *Element) *Element {
	var yInv Element
	yInv.Inverse(y)
	z.Mul(x, &yInv)
	return z
}

// Equal returns z == x
func (z *Element) Equal(x *Element) bool {
	return (z[1] == x[1]) && (z[0] == x[0])
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	return (z[1] | z[0]) == 0
}

// Cmp compares (lexicographic order) z and x and returns:
//
//   -1 if z <  x
//    0 if z == x
//   +1 if z >  x
//
func (z *Element) Cmp(x *Element) int {
	_z := *z
	_x := *x
	_z.FromMont()
	_x.FromMont()
	if _z[1] > _x[1] {
		return 1
	} else if _z[1] < _x[1] {
		return -1
	}
	if _z[0] > _x[0] {
		return 1
	} else if _z[0] < _x[0] {
		return -1
	}
	return 0
}

// LexicographicallyLargest returns true if this element is strictly lexicographically
// larger than its negation, false otherwise
func (z *Element) LexicographicallyLargest() bool {
	// adapted from github.com/zkcrypto/bls12_381
	// we check if the element is larger than (q-1) / 2
	// if z - (((q -1) / 2) + 1) have no underflow, then z > (q-1) / 2

	_z := *z
	_z.FromMont()

	var b uint64
	_, b = bits.Sub64(_z[0], 0, 0)
	_, b = bits.Sub64(_z[1], 4611686018427387904, b)

	return b == 0
}

// SetRandom sets z to a random element < q
func (z *Element) SetRandom(rnd io.Reader) (*Element, error) {
	var bytes [16]byte

	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}
	z[0] = binary.BigEndian.Uint64(bytes[0:8])
	z[1] = binary.BigEndian.Uint64(bytes[8:16])
	z[1] %= 9223372036854775807

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
	}

	return z, nil
}

func RandomVector(length int, rnd io.Reader) ([]Element, error) {
	bytesLength := length*16 + 1
	bytes := make([]byte, bytesLength)
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, err
	}
	zs := make([]Element, length)
	for i := 0; i < length; i++ {
		var z Element
		z[0] = binary.BigEndian.Uint64(bytes[0:8])
		z[1] = binary.BigEndian.Uint64(bytes[8:16])
		z[1] %= 9223372036854775807

		// if z > q --> z -= q
		// note: this is NOT constant time
		if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
			var b uint64
			z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
			z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
		}
		zs[i] = z
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

// One returns 1 (in montgommery form)
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

// MulAssign is deprecated
// Deprecated: use Mul instead
func (z *Element) MulAssign(x *Element) *Element {
	return z.Mul(z, x)
}

// AddAssign is deprecated
// Deprecated: use Add instead
func (z *Element) AddAssign(x *Element) *Element {
	return z.Add(z, x)
}

// SubAssign is deprecated
// Deprecated: use Sub instead
func (z *Element) SubAssign(x *Element) *Element {
	return z.Sub(z, x)
}

// API with assembly impl

// Mul z = x * y mod q
// see https://hackmd.io/@zkteam/modular_multiplication
func (z *Element) Mul(x, y *Element) *Element {
	mul(z, x, y)
	return z
}

// Square z = x * x mod q
// see https://hackmd.io/@zkteam/modular_multiplication
func (z *Element) Square(x *Element) *Element {
	square(z, x)
	return z
}

// FromMont converts z in place (i.e. mutates) from Montgomery to regular representation
// sets and returns z = z * 1
func (z *Element) FromMont() *Element {
	fromMont(z)
	return z
}

// Add z = x + y mod q
func (z *Element) Add(x, y *Element) *Element {
	add(z, x, y)
	return z
}

// Double z = x + x mod q, aka Lsh 1
func (z *Element) Double(x *Element) *Element {
	double(z, x)
	return z
}

// Sub  z = x - y mod q
func (z *Element) Sub(x, y *Element) *Element {
	sub(z, x, y)
	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	neg(z, x)
	return z
}

// Generic (no ADX instructions, no AMD64) versions of multiplication and squaring algorithms

func _mulGeneric(z, x, y *Element) {

	var t [3]uint64
	var D uint64
	var m, C uint64
	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(y[0], x[0])
	C, t[1] = madd1(y[0], x[1], C)

	D = C

	// m = t[0]n'[0] mod W
	m = t[0] * 1

	// -----------------------------------
	// Second loop
	C = madd0(m, 18446744073709551615, t[0])

	C, t[0] = madd3(m, 9223372036854775807, t[1], C, t[2])

	t[1], t[2] = bits.Add64(D, C, 0)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(y[1], x[0], t[0])
	C, t[1] = madd2(y[1], x[1], t[1], C)

	D = C

	// m = t[0]n'[0] mod W
	m = t[0] * 1

	// -----------------------------------
	// Second loop
	C = madd0(m, 18446744073709551615, t[0])

	C, t[0] = madd3(m, 9223372036854775807, t[1], C, t[2])

	t[1], t[2] = bits.Add64(D, C, 0)

	if t[2] != 0 {
		// we need to reduce, we have a result on 3 words
		var b uint64
		z[0], b = bits.Sub64(t[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(t[1], 9223372036854775807, b)

		return

	}

	// copy t into z
	z[0] = t[0]
	z[1] = t[1]

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
	}
}

func _squareGeneric(z, x *Element) {

	var t [3]uint64
	var D uint64
	var m, C uint64
	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], x[0])
	C, t[1] = madd1(x[0], x[1], C)

	D = C

	// m = t[0]n'[0] mod W
	m = t[0] * 1

	// -----------------------------------
	// Second loop
	C = madd0(m, 18446744073709551615, t[0])

	C, t[0] = madd3(m, 9223372036854775807, t[1], C, t[2])

	t[1], t[2] = bits.Add64(D, C, 0)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], x[0], t[0])
	C, t[1] = madd2(x[1], x[1], t[1], C)

	D = C

	// m = t[0]n'[0] mod W
	m = t[0] * 1

	// -----------------------------------
	// Second loop
	C = madd0(m, 18446744073709551615, t[0])

	C, t[0] = madd3(m, 9223372036854775807, t[1], C, t[2])

	t[1], t[2] = bits.Add64(D, C, 0)

	if t[2] != 0 {
		// we need to reduce, we have a result on 3 words
		var b uint64
		z[0], b = bits.Sub64(t[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(t[1], 9223372036854775807, b)

		return

	}

	// copy t into z
	z[0] = t[0]
	z[1] = t[1]

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
	}
}

func _fromMontGeneric(z *Element) {
	// the following lines implement z = z * 1
	// with a modified CIOS montgomery multiplication
	{
		// m = z[0]n'[0] mod W
		m := z[0] * 1
		C := madd0(m, 18446744073709551615, z[0])
		C, z[0] = madd2(m, 9223372036854775807, z[1], C)
		z[1] = C
	}
	{
		// m = z[0]n'[0] mod W
		m := z[0] * 1
		C := madd0(m, 18446744073709551615, z[0])
		C, z[0] = madd2(m, 9223372036854775807, z[1], C)
		z[1] = C
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
	}
}

func _addGeneric(z, x, y *Element) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	// if we overflowed the last addition, z >= q
	// if z >= q, z = z - q
	if carry != 0 {
		// we overflowed, so z >= q
		z[0], carry = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], carry = bits.Sub64(z[1], 9223372036854775807, carry)
		return
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
	}
}

func _doubleGeneric(z, x *Element) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], x[0], 0)
	z[1], carry = bits.Add64(x[1], x[1], carry)
	// if we overflowed the last addition, z >= q
	// if z >= q, z = z - q
	if carry != 0 {
		// we overflowed, so z >= q
		z[0], carry = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], carry = bits.Sub64(z[1], 9223372036854775807, carry)
		return
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
	}
}

func _subGeneric(z, x, y *Element) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], b = bits.Sub64(x[1], y[1], b)
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Add64(z[1], 9223372036854775807, c)
	}
}

func _negGeneric(z, x *Element) {
	if x.IsZero() {
		z.SetZero()
		return
	}
	var borrow uint64
	z[0], borrow = bits.Sub64(18446744073709551615, x[0], 0)
	z[1], _ = bits.Sub64(9223372036854775807, x[1], borrow)
}

func _reduceGeneric(z *Element) {

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[1] < 9223372036854775807 || (z[1] == 9223372036854775807 && (z[0] < 18446744073709551615))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744073709551615, 0)
		z[1], _ = bits.Sub64(z[1], 9223372036854775807, b)
	}
}

// Exp z = x^exponent mod q
func (z *Element) Exp(x Element, exponent *big.Int) *Element {
	var bZero big.Int
	if exponent.Cmp(&bZero) == 0 {
		return z.SetOne()
	}

	z.Set(&x)

	for i := exponent.BitLen() - 2; i >= 0; i-- {
		z.Square(z)
		if exponent.Bit(i) == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

// ToMont converts z to Montgomery form
// sets and returns z = z * r^2
func (z *Element) ToMont() *Element {
	return z.Mul(z, &rSquare)
}

// ToRegular returns z in regular form (doesn't mutate z)
func (z Element) ToRegular() Element {
	return *z.FromMont()
}

// String returns the string form of an Element in Montgomery form
func (z *Element) String() string {
	vv := bigIntPool.Get().(*big.Int)
	defer bigIntPool.Put(vv)
	return z.ToBigIntRegular(vv).String()
}

// ToBigInt returns z as a big.Int in Montgomery form
func (z *Element) ToBigInt(res *big.Int) *big.Int {
	var b [Limbs * 8]byte
	binary.BigEndian.PutUint64(b[8:16], z[0])
	binary.BigEndian.PutUint64(b[0:8], z[1])

	return res.SetBytes(b[:])
}

// ToBigIntRegular returns z as a big.Int in regular form
func (z Element) ToBigIntRegular(res *big.Int) *big.Int {
	z.FromMont()
	return z.ToBigInt(res)
}

// Bytes returns the regular (non montgomery) value
// of z as a big-endian byte array.
func (z *Element) Bytes() (res [Limbs * 8]byte) {
	_z := z.ToRegular()
	binary.BigEndian.PutUint64(res[8:16], _z[0])
	binary.BigEndian.PutUint64(res[0:8], _z[1])

	return
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value (in Montgomery form), and returns z.
func (z *Element) SetBytes(e []byte) *Element {
	// get a big int from our pool
	vv := bigIntPool.Get().(*big.Int)
	vv.SetBytes(e)

	// set big int
	z.SetBigInt(vv)

	// put temporary object back in pool
	bigIntPool.Put(vv)

	return z
}

// SetBigInt sets z to v (regular form) and returns z in Montgomery form
func (z *Element) SetBigInt(v *big.Int) *Element {
	z.SetZero()

	var zero big.Int

	// fast path
	c := v.Cmp(&_modulus)
	if c == 0 {
		// v == 0
		return z
	} else if c != 1 && v.Cmp(&zero) != -1 {
		// 0 < v < q
		return z.setBigInt(v)
	}

	// get temporary big int from the pool
	vv := bigIntPool.Get().(*big.Int)

	// copy input + modular reduction
	vv.Set(v)
	vv.Mod(v, &_modulus)

	// set big int byte value
	z.setBigInt(vv)

	// release object into pool
	bigIntPool.Put(vv)
	return z
}

// setBigInt assumes 0 <= v < q
func (z *Element) setBigInt(v *big.Int) *Element {
	vBits := v.Bits()

	if bits.UintSize == 64 {
		for i := 0; i < len(vBits); i++ {
			z[i] = uint64(vBits[i])
		}
	} else {
		for i := 0; i < len(vBits); i++ {
			if i%2 == 0 {
				z[i/2] = uint64(vBits[i])
			} else {
				z[i/2] |= uint64(vBits[i]) << 32
			}
		}
	}

	return z.ToMont()
}

// SetString creates a big.Int with s (in base 10) and calls SetBigInt on z
func (z *Element) SetString(s string) *Element {
	// get temporary big int from the pool
	vv := bigIntPool.Get().(*big.Int)

	if _, ok := vv.SetString(s, 10); !ok {
		panic("Element.SetString failed -> can't parse number in base10 into a big.Int")
	}
	z.SetBigInt(vv)

	// release object into pool
	bigIntPool.Put(vv)

	return z
}

var (
	_bLegendreExponentElement *big.Int
	_bSqrtExponentElement     *big.Int
)

func init() {
	_bLegendreExponentElement, _ = new(big.Int).SetString("3fffffffffffffffffffffffffffffff", 16)
	const sqrtExponentElement = "20000000000000000000000000000000"
	_bSqrtExponentElement, _ = new(big.Int).SetString(sqrtExponentElement, 16)
}

// Legendre returns the Legendre symbol of z (either +1, -1, or 0.)
func (z *Element) Legendre() int {
	var l Element
	// z^((q-1)/2)
	l.Exp(*z, _bLegendreExponentElement)

	if l.IsZero() {
		return 0
	}

	// if l == 1
	if (l[1] == 0) && (l[0] == 2) {
		return 1
	}
	return -1
}

// Sqrt z = √x mod q
// if the square root doesn't exist (x is not a square mod q)
// Sqrt leaves z unchanged and returns nil
func (z *Element) Sqrt(x *Element) *Element {
	// q ≡ 3 (mod 4)
	// using  z ≡ ± x^((p+1)/4) (mod q)
	var y, square Element
	y.Exp(*x, _bSqrtExponentElement)
	// as we didn't compute the legendre symbol, ensure we found y such that y * y = x
	square.Square(&y)
	if square.Equal(x) {
		return z.Set(&y)
	}
	return nil
}

// Inverse z = x^-1 mod q
// note: allocates a big.Int (math/big)
func (z *Element) Inverse(x *Element) *Element {
	var _xNonMont big.Int
	x.ToBigIntRegular(&_xNonMont)
	_xNonMont.ModInverse(&_xNonMont, Modulus())
	z.SetBigInt(&_xNonMont)
	return z
}