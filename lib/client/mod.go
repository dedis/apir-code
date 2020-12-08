package client

import (
	"errors"
	"math/rand"

	"github.com/ncw/gmp"
	"github.com/si-co/vpir-code/lib/constants"
	"golang.org/x/crypto/blake2b"
)

// Client represents the client instance in both the IT and C models
type Client interface {
	Query()
	Reconstruct()
}

// Information-theoretic PIR client implements the Client interface
type ITClient struct {
	xof blake2b.XOF
	state *itClientState
}

type itClientState struct {
	i     int
	alpha *gmp.Int
}

func NewITClient(xof blake2b.XOF) *ITClient {
	return &ITClient{
		xof: xof,
		state: nil,
	}
}

func (c *ITClient) Query(index int, numServers int) [][]*gmp.Int {
	if index < 0 || index > constants.DBLength {
		panic("query index out of bound")
	}
	if numServers < 1 {
		panic("need at least 1 server")
	}

	// sample random alpha
	alpha := new(gmp.Int)
	rnd := rand.New(rand.NewSource(777777))
	alpha.Rand(rnd, constants.Modulo)

	// set ITClient state
	c.state = &itClientState{i: index, alpha: alpha}

	// sample k (variable Servers) random vectors q0,..., q_{k-1} such
	// that they sum to alpha * e_i
	eialpha := make([]*gmp.Int, constants.DBLength)
	vectors := make([][]*gmp.Int, numServers)
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]*gmp.Int, constants.DBLength)
	}

	for i := 0; i < constants.DBLength; i++ {
		// create basic vector
		eialpha[i] = gmp.NewInt(0)

		// set alpha at the index we want to retrieve
		if i == index {
			eialpha[i] = alpha
		}

		// create k - 1 random vectors
		sum := gmp.NewInt(0)
		for k := 0; k < numServers-1; k++ {
			randInt := new(gmp.Int)
			randInt.Rand(rnd, constants.Modulo)
			vectors[k][i] = randInt
			sum.Add(sum, randInt)
		}
		vectors[numServers-1][i] = gmp.NewInt(0)
		vectors[numServers-1][i].Sub(eialpha[i], sum)
	}

	return vectors
}

func (c *ITClient) Reconstruct(answers []*gmp.Int) (*gmp.Int, error) {
	sum := gmp.NewInt(0)
	for _, a := range answers {
		sum.Add(sum, a)
	}

	switch {
	case sum.Cmp(c.state.alpha) == 0:
		return constants.BigOne, nil
	case sum.Cmp(constants.BigZero) == 0:
		return constants.BigZero, nil
	default:
		return nil, errors.New("REJECT!")
	}
}
