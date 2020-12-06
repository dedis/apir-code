package client

import (
	"crypto/rand"
	"errors"
	"math/big"
	
	cst "github.com/si-co/vpir-code/lib/constants"
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
	alpha *big.Int
}

func NewITClient(xof blake2b.XOF) *ITClient {
	return &ITClient{
		xof: xof,
		state: nil,
	}
}

func (c *ITClient) Query(index int, numServers int) [][]*big.Int {
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

	// set ITClient state
	c.state = &itClientState{i: index, alpha: alpha}

	// sample k (variable Servers) random vectors q0,..., q_{k-1} such
	// that they sum to alpha * e_i
	eialpha := make([]*big.Int, cst.DBLength)
	vectors := make([][]*big.Int, numServers)
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]*big.Int, cst.DBLength)
	}

	for i := 0; i < cst.DBLength; i++ {
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

func (c *ITClient) Reconstruct(answers []*big.Int) (*big.Int, error) {
	sum := big.NewInt(0)
	for _, a := range answers {
		sum.Add(sum, a)
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
