package fss

// Source: https://github.com/frankw2/libfss/blob/master/go/test_fss/test_fss.go

import (
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

const (
	testBlockLength = 17
	numBits         = 64
)

func TestPoint(t *testing.T) {
	// Generate fss Keys on client
	fClient := ClientInitialize(numBits, testBlockLength)

	// random index, biased but fine for this test
	index := randomIndex(numBits)

	// Test with if x = index, evaluate to vector b
	bLen := testBlockLength
	b := make([]uint32, bLen)
	for i := range b {
		b[i] = field.RandElement()
	}
	fssKeys := fClient.GenerateTreePF(index, b)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits, testBlockLength)

	zeros := make([]uint32, bLen)
	// test only part of the input space, impossible to do a complete test
	// over 64 bits
	for j := 0; j <= 10000; j++ {
		indexToTest := randomIndex(numBits)
		if j == 0 {
			indexToTest = index
		}
		out0 := make([]uint32, bLen)
		out1 := make([]uint32, bLen)
		sum := make([]uint32, bLen)

		fServer.EvaluatePF(0, fssKeys[0], indexToTest, out0)
		fServer.EvaluatePF(1, fssKeys[1], indexToTest, out1)

		for i := range sum {
			sum[i] = (out0[i] + out1[i]) % field.ModP
		}

		if equalIndices(index, indexToTest) {
			require.Equal(t, b, sum)
		} else {
			require.Equal(t, zeros, sum)
		}
	}
}

func TestPointWithAlphaVector(t *testing.T) {
	// Generate fss Keys on client
	fClient := ClientInitialize(numBits, testBlockLength)

	// random index, biased but fine for this test
	index := make([]bool, 64)
	for i := range index {
		index[i] = (rand.Intn(2) % 2) != 0
	}

	// Test with if x = index, evaluate to vector b
	alpha := field.RandElementWithPRG(utils.RandomPRG())
	bLen := testBlockLength
	b := make([]uint32, testBlockLength)
	b[0] = 1
	for i := 1; i < len(b); i++ {
		a := (uint64(b[i-1]) * uint64(alpha)) % uint64(field.ModP)
		b[i] = uint32(a)
	}

	fssKeys := fClient.GenerateTreePF(index, b)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits, testBlockLength)

	zeros := make([]uint32, bLen)
	// test only random samples of the input space, impossible to do a complete test
	// over 64 bits
	for j := 0; j <= 10000; j++ {
		indexToTest := randomIndex(numBits)
		if j == 0 {
			indexToTest = index
		}
		out0 := make([]uint32, bLen)
		out1 := make([]uint32, bLen)
		sum := make([]uint32, bLen)

		fServer.EvaluatePF(0, fssKeys[0], indexToTest, out0)
		fServer.EvaluatePF(1, fssKeys[1], indexToTest, out1)

		for i := range sum {
			sum[i] = (out0[i] + out1[i]) % field.ModP
		}

		if equalIndices(index, indexToTest) {
			require.Equal(t, b, sum)
		} else {
			require.Equal(t, zeros, sum)
		}
	}
}

func randomIndex(bits int) []bool {
	index := make([]bool, bits)
	for i := range index {
		index[i] = (rand.Intn(2) % 2) != 0
	}

	return index
}

func equalIndices(a, b []bool) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
