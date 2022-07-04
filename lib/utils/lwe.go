package utils

import "crypto/aes"

// WARNING: DO NOT USE THESE KEYS IN PRODUCTION!

var SeedMatrixA = [aes.BlockSize]byte{19, 177, 222, 148, 155, 239, 159, 227, 155, 99, 246, 214, 220, 162, 30, 66}

func GetDefaultSeedMatrixA() *PRGKey {
	key := PRGKey(SeedMatrixA)
	return &key
}
