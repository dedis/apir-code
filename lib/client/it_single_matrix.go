package client

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"

	cst "github.com/si-co/vpir-code/lib/constants"
	"golang.org/x/crypto/blake2b"
)

// Information-theoretic PIR client implements the Client interface
type ITMatrixClient struct {
	xof   blake2b.XOF
	state *itMatrixClientState
}

type itMatrixClientState struct {
	ix    int
	iy    int
	alpha *big.Int
}

func NewITMatrixClient(xof blake2b.XOF) *ITMatrixClient {
	return &ITMatrixClient{
		xof:   xof,
		state: nil,
	}
}

func (c *ITMatrixClient) Query(index int, numServers int) [][]*big.Int {
	if index < 0 || index > cst.DBLength {
		panic("query index out of bound")
	}
	if numServers < 1 {
		panic("need at least 1 server")
	}

	// sample random alpha
	alpha, err := rand.Int(c.xof, cst.Modulo)
	if err != nil {
		panic(err)
	}

	// verified at server side if integer square
	dbLengthSqrt := int(math.Sqrt(cst.DBLength))

	// set ITClient state
	c.state = &itMatrixClientState{
		ix:    index / dbLengthSqrt,
		iy:    index % dbLengthSqrt,
		alpha: alpha,
	}

	eialpha := make([]*big.Int, dbLengthSqrt)
	vectors := make([][]*big.Int, numServers)
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]*big.Int, dbLengthSqrt)
	}

	for i := 0; i < dbLengthSqrt; i++ {
		// create basic vector
		eialpha[i] = big.NewInt(0)

		// set alpha at the index we want to retrieve
		if i == index {
			eialpha[i] = alpha
		}

		// create k - 1 random vectors
		sum := big.NewInt(0)
		for k := 0; k < numServers-1; k++ {
			randInt, err := rand.Int(c.xof, cst.Modulo)
			if err != nil {
				panic(err)
			}
			vectors[k][i] = randInt
			sum.Add(sum, randInt)
		}
		vectors[numServers-1][i] = new(big.Int)
		vectors[numServers-1][i].Sub(eialpha[i], sum)
	}

	return vectors

}

func (c *ITMatrixClient) Reconstruct(answers [][]*big.Int, rebalanced bool) (*big.Int, error) {
	sum := big.NewInt(0)
	for _, a := range answers {
		for _, singleAnswer := range a {
			sum.Add(sum, singleAnswer)
		}
	}

	switch {
	case sum.Cmp(c.state.alpha) == 0:
		return cst.BigOne, nil
	case sum.Cmp(cst.BigZero) == 0:
		return cst.BigZero, nil
	default:
		return nil, errors.New("REJECT!")
	}
}
