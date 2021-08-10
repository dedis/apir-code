package fss

// Source: https://github.com/frankw2/libfss/blob/master/go/test_fss/test_fss.go

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoint(t *testing.T) {
	// Generate fss Keys on client
	fClient := ClientInitialize(6)
	// Test with if x = 10, evaluate to vector b
	bLen := 100000
	b := make([]uint, bLen)
	for i := range b {
		b[i] = uint(rand.Int())
	}
	fssKeys := fClient.GenerateTreePF(10, b)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	out0 := make([]int, bLen)
	out1 := make([]int, bLen)
	sum := make([]uint, bLen)

	fServer.EvaluatePF(0, fssKeys[0], 10, out1)
	fServer.EvaluatePF(1, fssKeys[1], 10, out0)
	for i := range sum {
		sum[i] = uint(out0[i]) + uint(out1[i])
	}
	require.Equal(t, b, sum)

	fServer.EvaluatePF(0, fssKeys[0], 1, out1)
	fServer.EvaluatePF(1, fssKeys[1], 1, out0)
	for i := range sum {
		sum[i] = uint(out0[i]) + uint(out1[i])
	}
	require.NotEqual(t, b, sum)

}
