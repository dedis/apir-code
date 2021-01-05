package field

// Mostly adapted from:
// https://golang.org/src/crypto/cipher/gcm.go

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"strconv"

	"golang.org/x/crypto/blake2b"

	our_rand "github.com/si-co/vpir-code/lib/utils"
)

type Element struct {
	low, high uint64
	//productTable [16]Element
}

type PrecompElement struct {
	productTable [16]*Element
}

func NewElement(in []byte) Element {
	if len(in) != 16 {
		panic("incorrect length")
	}

	low := binary.BigEndian.Uint64(in[:8])
	high := binary.BigEndian.Uint64(in[8:])

	return Element{
		low:  low,
		high: high,
	}
}

func Zero() Element {
  return Element{low:0, high:0}
}

func One() Element {
	one := Zero()
	// the coefficient of x⁰ can be obtained by v.low >> 63.
	one.low ^= (1 << 63)
	return one
}

// Generator of the multiplicative group
func Gen() Element {
	gen := Zero()
	// the coefficient of x^1 can be obtained by v.low >> 62.
	gen.low ^= (1 << 62)
	return gen
}

func FromBytes(in []byte) Element {
  if len(in) == 16 {
    return NewElement(in[:])
  }

  sum := blake2b.Sum256(in[:])
  return NewElement(sum[:16])
}

func RandomXOF(xof blake2b.XOF) Element {
	var bytes [16]byte
	_, err := io.ReadFull(xof, bytes[:])
	if err != nil {
		panic("Should never get here")
	}

	return NewElement(bytes[:])
}

func (e *Element) Negate() {
  // Does nothing in field of characteristic 2
}

func Random() Element {
	var bytes [16]byte
	_, err := rand.Read(bytes[:])
	if err != nil {
		panic("Should never get here")
	}

	return NewElement(bytes[:])
}

func RandomVectorXOF(length int, xof blake2b.XOF) []Element {
	bytesLength := length*16 + 1
	bytes := make([]byte, bytesLength)
	_, err := io.ReadFull(xof, bytes[:])
	if err != nil {
		panic("Should never get here")
	}
	elements := make([]Element, length)
	for i := 0; i < bytesLength-16; i += 16 {
		elements[i/16] = NewElement(bytes[i : i+16])
	}

	return elements
}

func RandomVectorPRG(length int, prg *our_rand.PRGReader) []Element {
	bytesLength := length*16 + 1
	bytes := make([]byte, bytesLength)
	_, err := prg.Read(bytes)
	if err != nil {
		panic("Should never get here")
	}
	elements := make([]Element, length)
	for i := 0; i < bytesLength-16; i += 16 {
		elements[i/16] = NewElement(bytes[i : i+16])
	}

	return elements
}

// Set y = e*y, where the precomputed table is powers of e
func mulPrecomp(y *Element, productTable [16]*Element) *Element {
  z := new(Element)

	for i := 0; i < 2; i++ {
		word := y.high
		if i == 1 {
			word = y.low
		}

		// Multiplication works by multiplying z by 16 and adding in
		// one of the precomputed multiples of H.
		for j := 0; j < 64; j += 4 {
			msw := z.high & 0xf
			z.high >>= 4
			z.high |= z.low << 60
			z.low >>= 4
			z.low ^= uint64(gcmReductionTable[msw]) << 48

			// the values in |table| are ordered for
			// little-endian bit positions. See the comment
			// in NewGCMWithNonceSize.
			t := productTable[word&0xf]

			z.low ^= t.low
			z.high ^= t.high
			word >>= 4
		}
	}

  return z
}

func Mul(x, y Element) Element {
	productTable := createProductTable(x)
  return *mulPrecomp(&y, productTable)
}

// Multiply other by precomputed element and store result in other
func (pe *PrecompElement) MulBy(other *Element) {
  *other = *mulPrecomp(other, pe.productTable)
}


func (e Element) PrecomputeMul() PrecompElement {
  return PrecompElement { productTable: createProductTable(e) }
}

func (e Element) Equal(x Element) bool {
	return e.high == x.high && e.low == x.low
}

func (e Element) String() string {
	return strconv.FormatUint(e.low, 16) + strconv.FormatUint(e.high, 16)
}

func (e Element) HexString() string {
	return hex.EncodeToString(e.Bytes())
}

func (e Element) Bytes() []byte {
	out := make([]byte, 16)
	binary.BigEndian.PutUint64(out[:8], e.low)
	binary.BigEndian.PutUint64(out[8:], e.high)

	return out
}

func createProductTable(e Element) [16]*Element {
  var productTable [16]*Element
  zero := Zero()
  productTable[0] = &zero
	productTable[reverseBits(1)] = &e

	for i := 2; i < 16; i += 2 {
    v1 := gcmMultiplyByH(*productTable[reverseBits(i/2)])
		productTable[reverseBits(i)] = &v1
    v2 := Add(*productTable[reverseBits(i)], e)
		productTable[reverseBits(i+1)] = &v2
	}

	return productTable
}

func (e *Element) AddTo(x *Element) {
	e.low ^= x.low
  e.high ^= x.high
}

func Add(x, y Element) Element {
	// Addition in a characteristic 2 field is just XOR.
	return Element{low: x.low ^ y.low, high: x.high ^ y.high}
}

// gcmMultiplyByH returns the result of multiplying an element of GF(2¹²⁸)
// by the element x.
func gcmMultiplyByH(x Element) (double Element) {
	msbSet := x.high&1 == 1

	// Because of the bit-ordering, doubling is actually a right shift.
	double.high = x.high >> 1
	double.high |= x.low << 63
	double.low = x.low >> 1

	// If the most-significant bit was set before shifting then it,
	// conceptually, becomes a term of x^128. This is greater than the
	// irreducible polynomial so the result has to be reduced. The
	// irreducible polynomial is 1+x+x^2+x^7+x^128. We can subtract that to
	// eliminate the term at x^128 which also means subtracting the other
	// four terms. In characteristic 2 fields, subtraction == addition ==
	// XOR.
	if msbSet {
		double.low ^= 0xe100000000000000
	}

	return
}

var gcmReductionTable = []uint16{
	0x0000, 0x1c20, 0x3840, 0x2460, 0x7080, 0x6ca0, 0x48c0, 0x54e0,
	0xe100, 0xfd20, 0xd940, 0xc560, 0x9180, 0x8da0, 0xa9c0, 0xb5e0,
}

// reverseBits reverses the order of the bits of 4-bit number in i.
func reverseBits(i int) int {
	i = ((i << 2) & 0xc) | ((i >> 2) & 0x3)
	i = ((i << 1) & 0xa) | ((i >> 1) & 0x5)
	return i
}
