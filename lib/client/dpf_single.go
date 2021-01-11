package client

import (
	"errors"
	"math/bits"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
	"golang.org/x/crypto/blake2b"
)

type DPF struct {
	xof   blake2b.XOF
	state *dpfState
}

type dpfState struct {
	i     int
	alpha field.Element
}

func NewDPF(xof blake2b.XOF) *DPF {
	return &DPF{
		xof:   xof,
		state: nil,
	}
}

func (c *DPF) Query(index int, numServers int) ([][]byte, []dpf.FssKeyEq2P) {
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
	alpha, err := new(field.Element).SetRandom(c.xof)
	if err != nil {
		panic(err)
	}

	// set ITClient state
	c.state = &dpfState{i: index, alpha: *alpha}

	fClient := dpf.ClientInitialize(uint(bits.Len(uint(constants.DBLength))))
	fssKeys := fClient.GenerateTreePF(uint(index), alpha)

	return fClient.PrfKeys, fssKeys

}

func (c *DPF) Reconstruct(answers [][]field.Element) ([]field.Element, error) {
	answersLen := len(answers[0])
	sum := make([]field.Element, answersLen)

	for i := 0; i < answersLen; i++ {
		sum[i] = field.Zero()
		for s := range answers {
			sum[i].Add(&sum[i], &answers[s][i])
		}

	}

	i := 0
	switch {
	case sum[i].Equal(&c.state.alpha):
		return []field.Element{cst.One}, nil
	case sum[i].Equal(&cst.Zero):
		return []field.Element{cst.Zero}, nil
	default:
		return nil, errors.New("REJECT!")
	}

}
