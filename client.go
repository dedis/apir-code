package main

import (
	"crypto/rand"
	"errors"
	"math/big"
)

type Client struct {
}

type clientState struct {
	i     int
	alpha *big.Int
}

func (c Client) Query(i int) ([][]*big.Int, clientState) {
	if i < 0 || i > DBLength {
		panic("query index out of bound")
	}

	// sample random alpha
	alpha, err := rand.Int(rand.Reader, Modulo)
	if err != nil {
		panic(err)
	}

	// set client state
	st := clientState{i: i, alpha: alpha}

	// sample k (variable Servers) random vectors q0,..., q_{k-1} such
	// that they sum to alpha * e_i
	eialpha := make([]*big.Int, DBLength)
	vectors := make([][]*big.Int, Servers)
	for k := 0; k < Servers; k++ {
		vectors[k] = make([]*big.Int, DBLength)
	}

	for n := 0; n < DBLength; n++ {
		// create basic vector
		eialpha[n] = big.NewInt(0)

		// set alpha at the index we want to retrieve
		if n == i {
			eialpha[n] = alpha
		}

		// create k - 1 random vectors
		sum := big.NewInt(0)
		for k := 0; k < Servers-1; k++ {
			randInt, err := rand.Int(rand.Reader, Modulo)
			if err != nil {
				panic(err)
			}
			vectors[k][n] = randInt
			sum.Add(sum, randInt)
		}
		vectors[Servers-1][n] = new(big.Int)
		vectors[Servers-1][n].Sub(eialpha[n], sum)
	}

	return vectors, st
}

func (c Client) Reconstruct(answers []*big.Int, st clientState) (*big.Int, error) {
	sum := big.NewInt(0)
	for _, a := range answers {
		sum.Add(sum, a)
	}

	if sum.Cmp(st.alpha) != 0 && sum.Cmp(bigZero) != 0 {
		return nil, errors.New("REJECT!")
	}

	if sum.Cmp(st.alpha) == 0 {
		return bigOne, nil
	}

	return bigZero, nil
}
