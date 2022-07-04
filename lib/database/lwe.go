package database

import (
	crand "crypto/rand"
	"io"
	"math"
	"math/big"
	"math/rand"
)

type LWE struct {
	Entries []uint32
	Info
}

const plaintextModulus = 2

func CreateZeroLWE(numRows, numColumns int) *LWE {
	blockSize := 1 // for backward compatibility
	entries := make([]uint32, numRows*numColumns)

	return &LWE{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: numColumns,
			BlockSize:  blockSize,
		},
	}
}

func CreateRandomBinaryLWE(rnd io.Reader, numRows, numColumns int) *LWE {
	out := CreateZeroLWE(numRows, numColumns)
	for i := 0; i < numRows; i++ {
		for j := 0; j < numColumns; j++ {
			// TODO LWE: Replace with something real
			val := uint32(3*uint32(i)+7*uint32(j)) % plaintextModulus
			if val >= plaintextModulus {
				panic("plaintext value too large")
			}
			out.Set(i, j, val)
		}
	}

	return out
}

func CreateRandomLWE(rnd io.Reader, numRows, numColumns int, mod uint64) *LWE {
	modBig := big.NewInt(int64(mod))

	// TODO LWE: Replace with something much faster
	m := CreateZeroLWE(numRows, numColumns)
	for i := 0; i < len(m.Entries); i++ {
		v, err := crand.Int(rnd, modBig)
		if err != nil {
			panic("Error reading random int")
		}
		m.Entries[i] = uint32(v.Uint64())
	}

	return m
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

func (m *LWE) Set(r int, c int, v uint32) {
	m.Entries[m.NumColumns*r+c] = v
}

func (m *LWE) Get(r int, c int) uint32 {
	return m.Entries[m.NumColumns*r+c]
}

func (m *LWE) Rows() int {
	return m.NumRows
}

func (m *LWE) Cols() int {
	return m.NumColumns
}

func Mul(a *LWE, b *LWE) *LWE {
	if a.NumColumns != b.NumRows {
		panic("Dimension mismatch")
	}

	// TODO LWE Implement this inner loop in C for performance
	out := CreateZeroLWE(a.NumRows, b.NumColumns)
	for i := 0; i < a.NumRows; i++ {
		for k := 0; k < a.NumColumns; k++ {
			for j := 0; j < b.NumColumns; j++ {
				out.Entries[b.NumColumns*i+j] += a.Entries[a.NumColumns*i+k] * b.Entries[b.NumColumns*k+j]
			}
		}
	}

	return out
}

func (a *LWE) Add(b *LWE) {
	if a.NumColumns != b.NumColumns || a.NumRows != b.NumRows {
		panic("Dimension mismatch")
	}

	for i := 0; i < len(a.Entries); i++ {
		a.Entries[i] += b.Entries[i]
	}
}

func (a *LWE) Sub(b *LWE) {
	if a.NumColumns != b.NumColumns || a.NumRows != b.NumRows {
		panic("Dimension mismatch")
	}

	for i := 0; i < len(a.Entries); i++ {
		a.Entries[i] -= b.Entries[i]
	}
}
