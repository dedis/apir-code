package client

import (
	"errors"
	"fmt"

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
func NewITMulti(xof blake2b.XOF, rebalanced bool) *ITMulti {
	return &ITMulti{
		xof:        xof,
		rebalanced: rebalanced,
		state:      nil,
	}
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITMulti) Query(index int, numServers int) [][][]field.Element {
	// TODO: check query inputs
	blockLength := constants.BlockLength

	// sample random alpha using blake2b
	alpha := field.RandomXOF(c.xof)

	// set state
	switch c.rebalanced {
	case false:
		// iy is unused if the database is represented as a vector
		c.state = &itMultiState{
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
	for i := 1; i < len(a); i++ {
		e := &alpha
		power := a[i-1].PrecomputeMul()
		power.MulBy(e)
		a[i] = *e
	}

	// additive secret sharing
	aExtended := []field.Element{field.One()}
	aExtended = append(aExtended, a...)

	eia := make([][]field.Element, c.state.dbLength)
	for i := range eia {
		eia[i] = make([]field.Element, 1+blockLength)
	}

	// create query vectors for all the servers
	vectors := make([][][]field.Element, numServers)
	for k := range vectors {
		vectors[k] = make([][]field.Element, c.state.dbLength)
		for i := range vectors[0] {
			vectors[k][i] = make([]field.Element, blockLength)
		}
	}

	// zero vector in GF(2^128)^(1+b)
	// TODO: can we use constant?
	zeroVector := make([]field.Element, 1+blockLength)
	for i := range zeroVector {
		zeroVector[i] = field.Zero()
	}

	// perform additive secret sharing
	// TODO: optimize!
	for i := 0; i < c.state.dbLength; i++ {
		// create basic vector
		eia[i] = zeroVector

		// set alpha at the index we want to retrieve
		if i == c.state.ix {
			// aExtended is a vector of 1+b field elements
			eia[i] = aExtended
		}

		// create k - 1 random vectors of length dbLength containing
		// elements in GF(2^128)^(1+b)
		for k := 0; k < numServers-1; k++ {
			rand := field.RandomVectorXOF(blockLength, c.xof)
			vectors[k][i] = rand
		}

		// we should perform component-wise additive secret sharing
		sum := field.Zero()
		for b := 0; b < blockLength; b++ {
			for k := 0; k < numServers-1; k++ {
				sum = field.Add(sum, vectors[k][i][b])
			}
			vectors[numServers-1][i][b] = field.Add(eia[i][b], sum)
		}
	}

	return vectors
}

func (c *ITMulti) Reconstruct(answers [][]field.Element) ([]field.Element, error) {
	answersLen := len(answers[0])
	sum := make([]field.Element, answersLen)

	// sum answers as vectors in GF(2^128)^(1+b)
	for i := 0; i < answersLen; i++ {
		sum[i] = field.Zero()
		for s := range answers {
			sum[i] = field.Add(sum[i], answers[s][i])
		}

	}

	tag := sum[len(sum)-1]
	messages := sum[:len(sum)-1]

	// compute reconstructed tag

	// compute vector a = (alpha, alpha^2, ..., alpha^b)
	// TODO: store this in the state to avoid recomputation
	a := make([]field.Element, constants.BlockLength)
	a[0] = c.state.alpha
	prod := field.Mul(a[0], messages[0])
	reconstructedTag := prod
	for i := 1; i < len(a); i++ {
		e := &c.state.alpha
		power := a[i-1].PrecomputeMul()
		power.MulBy(e)
		prod := field.Mul(*e, messages[i])
		reconstructedTag = field.Add(reconstructedTag, prod)
	}
	fmt.Println(messages)
	if !tag.Equal(reconstructedTag) {
		return nil, errors.New("REJECT")
	}

	return messages, nil
}
