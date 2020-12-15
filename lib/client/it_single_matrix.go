package client

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

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
	ix := index % dbLengthSqrt
	iy := index / dbLengthSqrt

	// set ITClient state
	c.state = &itMatrixClientState{
		ix:    ix,
		iy:    iy,
		alpha: alpha,
	}

	vectors, err := utils.AdditiveSecretSharing(alpha, cst.Modulo, ix, dbLengthSqrt, numServers, c.xof)
	if err != nil {
		panic(err)
	}

	return vectors

}

func (c *ITMatrixClient) Reconstruct(answers [][]*big.Int) (*big.Int, error) {
	sum := make([]*big.Int, len(answers[0]))
	for i := 0; i < len(answers[0]); i++ {
		sum[i] = big.NewInt(0)
		for s := range answers {
			sum[i].Add(sum[i], answers[s][i])
		}

		if sum[i].Cmp(c.state.alpha) != 0 && sum[i].Cmp(big.NewInt(0)) != 0 {
			return nil, errors.New("REJECT!")
		}

	}

	switch {
	case sum[c.state.iy].Cmp(c.state.alpha) == 0:
		return cst.BigOne, nil
	case sum[c.state.iy].Cmp(cst.BigZero) == 0:
		return cst.BigZero, nil
	default:
		return nil, errors.New("REJECT!")
	}
}
