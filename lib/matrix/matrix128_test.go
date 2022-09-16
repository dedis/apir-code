package matrix

import (
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	"lukechampine.com/uint128"
)

func TestMultiply128(t *testing.T) {
	rnd := utils.RandomPRG()
	rows := 4
	cols := 4

	// start by multiplying with zeros
	m := NewRandom128(rnd, rows, cols)
	for i := range m.data {
		m.data[i] = uint128.From64(uint64(4))
	}
	// b is initialized at zero everywhere
	b := New128(rows, cols)
	// exp is a matrix of zeros
	exp := New128(rows, cols)
	res := Mul128(m, b)
	require.Equal(t, exp, res)

	// multiply with ones
	bOnes := New128(rows, cols)
	for i := range bOnes.data {
		bOnes.data[i] = uint128.From64(1)
	}
	expOnes := New128(rows, cols)
	for i := range expOnes.data {
		expOnes.data[i] = uint128.From64(16)
	}
	resOnes := Mul128(m, bOnes)

	require.Equal(t, expOnes, resOnes)
}

func TestBinaryMultiply128(t *testing.T) {
	rnd := utils.RandomPRG()
	rows := 4
	cols := 4

	// start by multiplying with zeros
	m := NewRandom128(rnd, rows, cols)
	for i := range m.data {
		m.data[i] = uint128.From64(uint64(4))
	}
	// b is initialized at zero everywhere
	b := NewBytes(rows, cols)
	// exp is a matrix of zeros
	exp := New128(rows, cols)
	res := BinaryMul128(m, b)
	require.Equal(t, exp, res)

	// multiply with ones
	bOnes := NewBytes(rows, cols)
	for i := range bOnes.data {
		bOnes.data[i] = 1
	}
	expOnes := New128(rows, cols)
	for i := range expOnes.data {
		expOnes.data[i] = uint128.From64(16)
	}
	resOnes := BinaryMul128(m, bOnes)

	require.Equal(t, expOnes, resOnes)

	// test if two routines behaves equally
	m1 := NewRandom128(rnd, rows, cols)
	b1 := NewBytes(rows, cols)
	for i := range b1.data {
		b1.data[i] = byte(i % 2)
	}
	b1_128 := New128(rows, cols)
	for i := range b1_128.data {
		b1_128.data[i] = uint128.From64(uint64(i % 2))
	}

	require.Equal(t, BinaryMul128(m1, b1), Mul128(m1, b1_128))
}

func TestMatrix128ToBytes(t *testing.T) {
	rows := rand.Intn(500)
	cols := rand.Intn(600)
	m := NewRandom128(utils.RandomPRG(), rows, cols)
	b := Matrix128ToBytes(m)
	require.Equal(t, m, BytesToMatrix128(b))
}
