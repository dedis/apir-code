package fss

// Source: https://github.com/frankw2/libfss/blob/master/go/test_fss/test_fss.go

import (
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/field"
	"github.com/stretchr/testify/require"
)

const testBlockLength = 16

func TestPoint(t *testing.T) {
	// Generate fss Keys on client
	fClient := ClientInitialize(6, testBlockLength)
	// Test with if x = 10, evaluate to vector b
	bLen := testBlockLength
	b := make([]uint32, bLen)
	for i := range b {
		b[i] = field.RandElement()
	}
	fssKeys := fClient.GenerateTreePF(10, b)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	out0 := make([]uint32, bLen)
	out1 := make([]uint32, bLen)
	sum := make([]uint32, bLen)

	fServer.EvaluatePF(0, fssKeys[0], 10, out0)
	fServer.EvaluatePF(1, fssKeys[1], 10, out1)
	for i := range sum {
		sum[i] = (out0[i] + out1[i]) % field.ModP
	}
	require.Equal(t, b, sum)

	fServer.EvaluatePF(0, fssKeys[0], 1, out0)
	fServer.EvaluatePF(1, fssKeys[1], 1, out1)
	for i := range sum {
		sum[i] = (out0[i] + out1[i]) % field.ModP
	}
	require.NotEqual(t, b, sum)
}

func TestInterval(t *testing.T) {
	// Generate fss Keys on client
	fClient := ClientInitialize(6, testBlockLength)
	// Test with if x < 10, evaluate to vector b
	bLen := testBlockLength
	b := make([]uint32, bLen)
	for i := range b {
		b[i] = rand.Uint32() % field.ModP
	}
	fssKeys := fClient.GenerateTreeLt(10, b)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	sum := make([]uint32, bLen)

	out0 := fServer.EvaluateLt(fssKeys[0], 1)
	out1 := fServer.EvaluateLt(fssKeys[1], 1)
	for i := range sum {
		sum[i] = (out0[i] + out1[i]) % field.ModP
	}
	require.Equal(t, b, sum)

	out0 = fServer.EvaluateLt(fssKeys[0], 11)
	out1 = fServer.EvaluateLt(fssKeys[1], 11)
	for i := range sum {
		sum[i] = (out0[i] + out1[i]) % field.ModP
	}
	require.NotEqual(t, b, sum)
}