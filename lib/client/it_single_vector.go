package client

import (
	"crypto/rand"
	"errors"
	"math/big"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

// Information-theoretic PIR client implements the Client interface
type ITClient struct {
	xof   blake2b.XOF
	state *itClientState
}

type itClientState struct {
	i     int
	alpha *big.Int
}

func NewITClient(xof blake2b.XOF) *ITClient {
	return &ITClient{
		xof:   xof,
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

	vectors, err := utils.AdditiveSecretSharing(alpha, cst.Modulo, index, cst.DBLength, numServers, c.xof)
	if err != nil {
		panic(err)
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
