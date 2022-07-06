package utils

import (
	"crypto/aes"

	"lukechampine.com/uint128"
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

	SeedA  *PRGKey // matrix  used to generate digest
	Mod    uint64
	Mod128 uint128.Uint128 // TODO: remove if we go with the 32-bits version
	Bytes  int
}

func ParamsDefault() *ParamsLWE {
	return &ParamsLWE{
		P:     2,
		N:     1024,
		Sigma: 6.0,
		L:     512,
		M:     128,
		B:     1000,
		SeedA: GetDefaultSeedMatrixA(),
		Mod:   1 << 32,
		Bytes: 4,
	}
}

func ParamsWithDatabaseSize(rows, columns int) *ParamsLWE {
	p := ParamsDefault()
	p.L = rows
	p.M = columns

	return p
}

// TODO: remove if we go with the 32-bits version
func ParamsDefault128() *ParamsLWE {
	p := ParamsDefault()
	p.Mod128 = uint128.Max
	p.Bytes = 16

	return p
}

func ParamsWithDatabaseSize128(rows, columns int) *ParamsLWE {
	p := ParamsDefault128()
	p.L = rows
	p.M = columns

	return p
}

func GetDefaultSeedMatrixA() *PRGKey {
	key := PRGKey(SeedMatrixA)
	return &key
}
