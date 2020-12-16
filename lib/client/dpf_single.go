package client

import (
	"crypto/rand"
	"errors"
	"math/big"
	"math/bits"

	"github.com/frankw2/libfss/go/libfss"
	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"golang.org/x/crypto/blake2b"
)

type DPF struct {
	xof   blake2b.XOF
	state *dpfState
}

type dpfState struct {
	i     int
	alpha *big.Int
}

func NewDPF(xof blake2b.XOF) *DPF {
	return &DPF{
		xof:   xof,
		state: nil,
	}
}

func (c *DPF) Query(index int, numServers int) ([][]byte, []libfss.FssKeyEq2P) {
	if index < 0 || index > cst.DBLength {
		panic("query index out of bound")
	}
	if numServers < 1 {
		panic("need at least 1 server")
	}
	if numServers != 2 {
		panic("DPF implementation only works with 2 servers")
	}

	// sample random alpha
	alpha, err := rand.Int(c.xof, cst.Modulo)
	if err != nil {
		panic(err)
	}
	alpha = big.NewInt(12)

	// set ITClient state
	c.state = &dpfState{i: index, alpha: alpha}

	fClient := libfss.ClientInitialize(uint(bits.Len(uint(constants.DBLength))))
	fssKeys := fClient.GenerateTreePF(uint(index), uint(alpha.Uint64()))

	return fClient.PrfKeys, fssKeys

}

func (c *DPF) Reconstruct(answers []*big.Int) (*big.Int, error) {
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
