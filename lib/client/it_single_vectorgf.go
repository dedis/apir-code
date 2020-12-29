package client

import (
	"errors"

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

	for i := 0; i < cst.DBLength; i++ {
		// create basic vector
		eialpha[i] = field.Zero()

		// set alpha at the index we want to retrieve
		if i == index {
			eialpha[i] = alpha
		}

		// create k - 1 random vectors
		sum := field.Zero()
		for k := 0; k < numServers-1; k++ {
			rand := field.RandomXOF(c.xof)
			vectors[k][i] = rand
			sum = field.Add(sum, rand)
		}
		vectors[numServers-1][i] = field.Add(eialpha[i], sum)
	}

	return vectors

}

func (c *ITVectorGF) Reconstruct(answers []*field.Element) (*field.Element, error) {
	sum := field.Zero()
	for _, a := range answers {
		sum = field.Add(sum, a)
	}

	switch {
	case sum.Equal(c.state.alpha):
		return field.One(), nil
	case sum.Equal(field.Zero()):
		return field.Zero(), nil
	default:
		return nil, errors.New("REJECT!")
	}

}
