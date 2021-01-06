package client

import (
	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"golang.org/x/crypto/blake2b"
)

// Information theoretic client for multi-bit scheme working in
// GF(2^128). Both vector and matrix (rebalanced) representations of
// the database are handled by this client, via a boolean variable

// ITMulti represents the client for the information theoretic multi-bit scheme
type ITMulti struct {
	xof        blake2b.XOF
	state      *itMultiState
	rebalanced bool
}

type itMultiState struct {
	ix       int
	iy       int // unused if not rebalanced
	alpha    field.Element
	dbLength int
}

// NewITSingleGF return a client for the information theoretic multi-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITMulti(xof blake2b.XOF, rebalanced bool) *ITSingleGF {
	return &ITSingleGF{
		xof:        xof,
		rebalanced: rebalanced,
		state:      nil,
	}
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITMulti) Query(index int, numServers int) [][]field.Element {
	// TODO: check query inputs
	blockLength = constants.BlockLength

	// sample random alpha using blake2b
	alpha := field.RandomXOF(c.xof)
	alphaPrecomp := alpha.PrecomputeMul()

	// set state
	switch c.rebalanced {
	case false:
		// iy is unused if the database is represented as a vector
		c.state = &itSingleGFState{
			ix:       index,
			alpha:    alpha,
			dbLength: cst.DBLength,
		}
	case true:
		panic("not yet implemented")
	}

	// compute vector a = (alpha, alpha^2, ..., alpha^b)
	// TODO: simplify field API
	a := make([]field.Element, blockLength)
	a[0] = alpha
	for i := range a[1:] {
		e := &alpha
		power := a[i-1].PrecomputeMul()
		power.MulBy(e)
		a[i] = *e
	}

	// additive secret sharing
	eialpha := make([]field.Element, c.state.dbLength*(1+blockLength))
	vectors := make([][]field.Element, numServers)

	// create query vectors for all the servers
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]field.Element, c.state.dbLength*(1+blockLength))
	}

	// zero vector in GF(2^128)^(1+b)
	zeroVector := make([]field.Element, 1+blockLength)
	for i := range zeroVector {
		// TODO: can we use constant?
		zeroVector[i] = field.Zero()
	}

	//numRandomElements := c.state.dbLength * (numServers - 1)
	//randomElements := field.RandomVectorXOF(numRandomElements, c.xof)
	for i := 0; i < c.state.dbLength; i++ {
		// create basic vector
		eialpha[i] = zero

		// set alpha at the index we want to retrieve
		if i == c.state.ix {
			eialpha[i] = c.state.alpha
		}

		// create k - 1 random vectors
		sum := field.Zero()
		for k := 0; k < numServers-1; k++ {
			rand := randomElements[c.state.dbLength*k+i]
			vectors[k][i] = rand
			sum = field.Add(sum, rand)
		}
		vectors[numServers-1][i] = field.Add(eialpha[i], sum)
	}

}
