package client

import (
	"crypto/rand"
	"errors"
	"math/big"

	cst "github.com/si-co/vpir-code/lib/constants"
	"golang.org/x/crypto/blake2b"
)

type client struct {
	xof blake2b.XOF
}

type clientState struct {
	i     int
	alpha *big.Int
}

func NewClient(xof blake2b.XOF) client {
	return client{xof: xof}
}

func (c client) Query(i int, servers int) ([][]*big.Int, clientState) {
	if i < 0 || i > cst.DBLength {
		panic("query index out of bound")
	}
	if servers < 1 {
		panic("need at least 1 server")
	}

	// sample random alpha
	alpha, err := rand.Int(rand.Reader, cst.Modulo)
	if err != nil {
		panic(err)
	}

	// set client state
	st := clientState{i: i, alpha: alpha}

	// sample k (variable Servers) random vectors q0,..., q_{k-1} such
	// that they sum to alpha * e_i
	eialpha := make([]*big.Int, cst.DBLength)
	vectors := make([][]*big.Int, servers)
	for k := 0; k < servers; k++ {
		vectors[k] = make([]*big.Int, cst.DBLength)
	}

	for n := 0; n < cst.DBLength; n++ {
		// create basic vector
		eialpha[n] = big.NewInt(0)

		// set alpha at the index we want to retrieve
		if n == i {
			eialpha[n] = alpha
		}

		// create k - 1 random vectors
		sum := big.NewInt(0)
		for k := 0; k < servers-1; k++ {
			randInt, err := rand.Int(c.xof, cst.Modulo)
			if err != nil {
				panic(err)
			}
			vectors[k][n] = randInt
			sum.Add(sum, randInt)
		}
		vectors[servers-1][n] = new(big.Int)
		vectors[servers-1][n].Sub(eialpha[n], sum)
	}

	return vectors, st
}

func (c client) Reconstruct(answers []*big.Int, st clientState) (*big.Int, error) {
	sum := big.NewInt(0)
	for _, a := range answers {
		sum.Add(sum, a)
	}

	switch {
	case sum.Cmp(st.alpha) == 0:
		return cst.BigOne, nil
	case sum.Cmp(cst.BigZero) == 0:
		return cst.BigZero, nil
	default:
		return nil, errors.New("REJECT!")
	}
}

func BitStringToBytes(s string) ([]byte, error) {
	b := make([]byte, (len(s)+(8-1))/8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '1' {
			return nil, errors.New("not a bit")
		}
		b[i>>3] |= (c - '0') << uint(7-i&7)
	}
	return b, nil
}
