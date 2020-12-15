package client

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/dimakogan/dpf-go/dpf"
	cst "github.com/si-co/vpir-code/lib/constants"
	"golang.org/x/crypto/blake2b"
)

type DPFClient struct {
	xof   blake2b.XOF
	state *itClientState
}

type dpfClientState struct {
	i     int
	alpha *big.Int
}

func NewDPFClient(xof blake2b.XOF) *DPFClient {
	return &DPFClient{
		xof:   xof,
		state: nil,
	}
}

func (c *DPFClient) Query(index int, numServers int) (dpf.DPFkey, dpf.DPFkey) {
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
	c.state = &itClientState{i: index, alpha: alpha}

	fmt.Println(uint64(math.Log2(cst.DBLength)))
	fmt.Println((1 << uint64(math.Log2(cst.DBLength))))
	fmt.Println(alpha.Uint64())
	return dpf.Gen(alpha.Uint64(), uint64(math.Log2(cst.DBLength)))

}

func (c *DPFClient) Reconstruct(answers []*big.Int) (*big.Int, error) {
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
