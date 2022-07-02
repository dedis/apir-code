package database

import (
	"math"
	"math/rand"
)

type LWE struct {
	Entries []uint32
	Info
}

func CreateZeroLWE(numRows, numColumns int) *LWE {
	blockSize = 1 // for backward compatibility
	entries := make([]uint32, numRows*numColumns)

	return &LWE{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: numColumns,
		},
	}
}

func CreateGaussLWe(numRows, numColumns int, sigma float64) *LWE {
	lwe := CreateZeroLWE(numRows, numColumns)
	for i := range lwe.Entries {
		lwe.Entries[i] = sampleGauss(sigma)
	}

	return lwe
}

func sampleGauss(sigma float64) uint32 {
	// TODO LWE: Replace with cryptographic RNG

	// Inspired by https://github.com/malb/dgs/
	tau := float64(18)
	upper_bound := int(math.Ceil(sigma * tau))
	f := -1.0 / (2.0 * sigma * sigma)

	x := 0
	for {
		// Sample random value in [-tau*sigma, tau-sigma]
		x = rand.Intn(2*upper_bound+1) - upper_bound
		diff := float64(x)
		accept_with_prob := math.Exp(diff * diff * f)
		if rand.Float64() < accept_with_prob {
			break
		}
	}

	return uint32(x)
}
