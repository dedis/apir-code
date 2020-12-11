package client

import (
	"crypto/rand"
	"errors"
	
	"github.com/holiman/uint256"
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
	alpha *uint256.Int
}

func NewITClient(xof blake2b.XOF) *ITClient {
	return &ITClient{
		xof: xof,
		state: nil,
	}
}

func (c *ITClient) Query(index int, numServers int) [][]*uint256.Int {
	if index < 0 || index > cst.DBLength {
		panic("query index out of bound")
	}
	if numServers < 1 {
		panic("need at least 1 server")
	}

	// sample random alpha
	alphaBig, err := rand.Int(c.xof, cst.Modulo)
	if err != nil {
		panic(err)
	}
	alpha, _ := uint256.FromBig(alphaBig)

	// set ITClient state
	c.state = &itClientState{i: index, alpha: alpha}

	// sample k (variable Servers) random vectors q0,..., q_{k-1} such
	// that they sum to alpha * e_i
	eialpha := make([]*uint256.Int, cst.DBLength)
	vectors := make([][]*uint256.Int, numServers)
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]*uint256.Int, cst.DBLength)
	}

	for i := 0; i < cst.DBLength; i++ {
		// create basic vector
		eialpha[i] = uint256.NewInt().SetUint64(0)

		// set alpha at the index we want to retrieve
		if i == index {
			eialpha[i] = alpha
		}

		// create k - 1 random vectors
		sum := uint256.NewInt().SetUint64(0)
		for k := 0; k < numServers-1; k++ {
			randIntBig, err := rand.Int(c.xof, cst.Modulo)
			if err != nil {
				panic(err)
			}
			randInt, _ := uint256.FromBig(randIntBig)
			vectors[k][i] = randInt
			sum.Add(sum, randInt)
		}
		vectors[numServers-1][i] = uint256.NewInt()
		vectors[numServers-1][i].Sub(eialpha[i], sum)
	}

	return vectors
}

func (c *ITClient) Reconstruct(answers []*uint256.Int) (*uint256.Int, error) {
	sum := uint256.NewInt().SetUint64(0)
	for _, a := range answers {
		sum.Add(sum, a)
	}

	switch {
	case sum.Cmp(c.state.alpha) == 0:
		return uint256.NewInt().SetUint64(1), nil
	case sum.Cmp(uint256.NewInt().SetUint64(0)) == 0:
		return uint256.NewInt().SetUint64(0), nil
	default:
		return nil, errors.New("REJECT!")
	}
}
