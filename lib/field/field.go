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
	"math/bits"
	"strconv"

	"golang.org/x/crypto/blake2b"

	our_rand "github.com/si-co/vpir-code/lib/utils"
)

// p = 2^64 - 59
const Modulus uint64 = 18446744073709551557
const NumBytes int = 8

type Element uint64

func NewElement(in []byte) Element {
	if len(in) != NumBytes {
		panic("incorrect length")
	}

  return reduce(Element(binary.BigEndian.Uint64(in[0:NumBytes])))
}

func reduce(e Element) Element {
  return Element(uint64(e) % Modulus)
}

func Zero() Element {
  return Element(0)
}

func One() Element {
  return Element(1)
}

func RandomXOF(xof blake2b.XOF) Element {
	var bytes [NumBytes]byte
	_, err := io.ReadFull(xof, bytes[:])
	if err != nil {
		panic("Should never get here")
	}

	return NewElement(bytes[:])
}

func (e *Element) Negate() {
  *e = Element(Modulus - uint64(*e))
}

func Random() Element {
	var bytes [NumBytes]byte
	_, err := rand.Read(bytes[:])
	if err != nil {
		panic("Should never get here")
	}

	return NewElement(bytes[:])
}

func RandomVectorXOF(length int, xof blake2b.XOF) []Element {
	bytesLength := length*NumBytes
	bytes := make([]byte, bytesLength)
	_, err := io.ReadFull(xof, bytes[:])
	if err != nil {
		panic("Should never get here")
	}
	elements := make([]Element, length)
  p := 0
	for i := 0; i < length; i += 1 {
		elements[i] = NewElement(bytes[p:p+NumBytes])
    p += NumBytes
	}

	return elements
}

func RandomVectorPRG(length int, prg *our_rand.PRGReader) []Element {
	bytesLength := length*NumBytes
	bytes := make([]byte, bytesLength)
	_, err := prg.Read(bytes)
	if err != nil {
		panic("Should never get here")
	}
	elements := make([]Element, length)
  p := 0
	for i := 0; i < length ; i += 1 {
		elements[i] = NewElement(bytes[p:p+NumBytes])
    p += NumBytes
	}

	return elements
}

// Multiply the two field elements
func Mul(e, y Element) Element {
  hi, low := bits.Mul64(uint64(e), uint64(y))
  return reduce(Element(bits.Rem64(hi, low, Modulus)))
}


func (e Element) Equal(x Element) bool {
	return uint64(e) == uint64(x)
}

func (e Element) String() string {
	return strconv.FormatUint(uint64(e), 16)
}

func (e Element) HexString() string {
	return hex.EncodeToString(e.Bytes())
}

func (e Element) Bytes() []byte {
	out := make([]byte, NumBytes)
	binary.BigEndian.PutUint64(out[:], uint64(e))
	return out
}

func (e Element) PrecomputeMul() {
  // Do nothing
}

func Add(x, y Element) Element {
  sum, carryOut := bits.Add64(uint64(x), uint64(y), 0)
  return Element(bits.Rem64(carryOut, sum, Modulus))
}

