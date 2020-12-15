package utils

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/blake2b"
)

// sample k (variable Servers) random vectors q0,..., q_{k-1} such
// that they sum to alpha * e_i
func AdditiveSecretSharing(alpha, modulo *big.Int, index, dbLength, numServers int, xof blake2b.XOF) ([][]*big.Int, error) {
	eialpha := make([]*big.Int, dbLength)
	vectors := make([][]*big.Int, numServers)
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]*big.Int, dbLength)
	}

	for i := 0; i < dbLength; i++ {
		// create basic vector
		eialpha[i] = big.NewInt(0)

		// set alpha at the index we want to retrieve
		if i == index {
			eialpha[i] = alpha
		}

		// create k - 1 random vectors
		sum := big.NewInt(0)
		for k := 0; k < numServers-1; k++ {
			randInt, err := rand.Int(xof, modulo)
			if err != nil {
				return nil, err
			}
			vectors[k][i] = randInt
			sum.Add(sum, randInt)
		}
		vectors[numServers-1][i] = new(big.Int)
		vectors[numServers-1][i].Sub(eialpha[i], sum)
	}

	return vectors, nil
}
