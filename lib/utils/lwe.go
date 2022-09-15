package utils

import (
	"crypto/aes"
	"math"
)

// WARNING: DO NOT USE THESE KEYS IN PRODUCTION!
var SeedMatrixA = [aes.BlockSize]byte{19, 177, 222, 148, 155, 239, 159, 227, 155, 99, 246, 214, 220, 162, 30, 66}

type ParamsLWE struct {
	P     uint32  // plaintext modulus
	N     int     // lattice/secret dimension
	Sigma float64 // Error parameter

	L int    // number of rows of database
	M int    // number of columns of database
	B uint32 // bound used in reconstruction

	SeedA    *PRGKey // matrix  used to generate digest
	BytesMod int     // bytes of the modulo
}

func ParamsDefault() *ParamsLWE {
	return &ParamsLWE{
		P:        2,
		N:        1100,
		Sigma:    6.4,
		SeedA:    GetDefaultSeedMatrixA(),
		BytesMod: 4,
	}
}

func ParamsWithDatabaseSize(rows, columns int) *ParamsLWE {
	p := ParamsDefault()
	p.L = rows
	p.M = columns
	p.B = computeB(rows, p.Sigma)

	return p
}

func GetDefaultSeedMatrixA() *PRGKey {
	key := PRGKey(SeedMatrixA)
	return &key
}

func ParamsDefault128() *ParamsLWE {
	p := ParamsDefault()
	p.N = 4800
	p.BytesMod = 16

	return p
}

func ParamsWithDatabaseSize128(rows, columns int) *ParamsLWE {
	p := ParamsDefault128()
	p.L = rows
	p.M = columns
	p.B = computeB(rows, p.Sigma)

	return p
}

func computeB(rows int, sigma float64) uint32 {
	// rows is equal to sqrt(dbSize), 12 is ~ sqrt(128)
	return uint32(rows * 12 * int(math.Ceil(sigma)))
}
