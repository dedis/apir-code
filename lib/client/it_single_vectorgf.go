package client

import (
	"errors"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"golang.org/x/crypto/blake2b"
)

// Information-theoretic PIR client implements the Client interface
type ITVectorGF struct {
	xof   blake2b.XOF
	state *itVectorGFState
}

type itVectorGFState struct {
	i     int
	alpha *field.Element
}

func NewITVectorGF(xof blake2b.XOF) *ITVectorGF {
	return &ITVectorGF{
		xof:   xof,
		state: nil,
	}
}

func (c *ITVectorGF) Query(index int, numServers int) [][]*field.Element {
	if index < 0 || index > cst.DBLength {
		panic("query index out of bound")
	}
	if numServers < 1 {
		panic("need at least 1 server")
	}

	alpha := field.RandomXOF(c.xof)

	// set ITVector state
	c.state = &itVectorGFState{i: index, alpha: alpha}

	eialpha := make([]*field.Element, cst.DBLength)
	vectors := make([][]*field.Element, numServers)
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]*field.Element, cst.DBLength)
	}

	zero := constants.Zero
	randomElements := field.RandomVectorXOF(cst.DBLength, c.xof)
	//randomElements := field.RandomVectorPRG(cst.DBLength, our_rand.RandomPRG())
	for i := 0; i < cst.DBLength; i++ {
		// create basic vector
		eialpha[i] = zero

		// set alpha at the index we want to retrieve
		if i == index {
			eialpha[i] = alpha
		}

		// create k - 1 random vectors
		sum := zero
		for k := 0; k < numServers-1; k++ {
			rand := randomElements[i]
			vectors[k][i] = rand
			sum.Add(sum, rand)
		}
		vectors[numServers-1][i] = field.Add(eialpha[i], sum)
	}

	return vectors
}

func (c *ITVectorGF) Reconstruct(answers []*field.Element) (*field.Element, error) {
	// TODO: here constants.Zero is not working, don't know why
	sum := field.Zero()
	for _, a := range answers {
		sum.Add(sum, a)
	}

	switch {
	case sum.Equal(c.state.alpha):
		return constants.One, nil
	case sum.Equal(constants.Zero):
		return constants.Zero, nil
	default:
		return nil, errors.New("REJECT!")
	}

}
