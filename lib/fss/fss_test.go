package fss

// Source: https://github.com/frankw2/libfss/blob/master/go/test_fss/test_fss.go

import (
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/stretchr/testify/require"
)

const testBlockLength = 16

func TestPoint(t *testing.T) {
	// Generate fss Keys on client
	fClient := ClientInitialize(6, testBlockLength)
	// Test with if x = 10, evaluate to vector b
	bLen := 1000
	b := make([]uint32, bLen)
	for i := range b {
		b[i] = rand.Uint32() % constants.ModP
	}
	fssKeys := fClient.GenerateTreePF(10, b)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	out0 := make([]uint32, bLen)
	out1 := make([]uint32, bLen)
	sum := make([]uint32, bLen)

	fServer.EvaluatePF(0, fssKeys[0], 10, out1)
	fServer.EvaluatePF(1, fssKeys[1], 10, out0)
	for i := range sum {
		sum[i] = (out0[i] + out1[i]) % constants.ModP
	}
	require.Equal(t, b, sum)

	fServer.EvaluatePF(0, fssKeys[0], 1, out1)
	fServer.EvaluatePF(1, fssKeys[1], 1, out0)
	for i := range sum {
		sum[i] = (out0[i] + out1[i]) % constants.ModP
	}
	require.NotEqual(t, b, sum)

}
