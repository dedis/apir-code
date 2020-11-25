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

func (c Client) Query(i int) ([]*big.Int, []*big.Int, clientState) {
	if i < 0 || i > DBLength {
		panic("query index out of bound")
	}

	// sample random alpha
	alpha, err := rand.Int(rand.Reader, MODULO)
	if err != nil {
		panic(err)
	}

	// set client state
	st := clientState{i: i, alpha: alpha}

	// sample two random vectors q0 and q1 such that q0 + q1 = eiy
	eialpha := make([]*big.Int, DBLength)
	q0 := make([]*big.Int, DBLength)
	q1 := make([]*big.Int, DBLength)

	for k := range q0 {
		// create basic vector
		eialpha[k] = bigZero

		if k == i {
			eialpha[k] = alpha
		}

		// create random vector
		randInt, err := rand.Int(rand.Reader, MODULO)
		if err != nil {
			panic(err)
		}
		q0[k] = randInt

		q1[k] = new(big.Int)
		q1[k].Sub(eialpha[k], q0[k])
	}

	return q0, q1, st
}

func (c Client) Reconstruct(a0, a1 *big.Int, st clientState) (*big.Int, error) {
	a := new(big.Int)
	a.Add(a0, a1)

	if a.Cmp(st.alpha) != 0 && a.Cmp(bigZero) != 0 {
		return nil, errors.New("REJECT!")
	}

	if a.Cmp(st.alpha) == 0 {
		return bigOne, nil
	}

	return bigZero, nil
}
