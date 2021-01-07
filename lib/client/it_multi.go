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
	a        []field.Element
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
	var alpha field.Element
	alpha.SetRandom(c.xof)

	// compute vector a = (alpha, alpha^2, ..., alpha^b)
	// TODO: simplify field API
	a := make([]field.Element, blockLength)
	a[0] = alpha
	for i := 1; i < len(a); i++ {
		a[i].Mul(&a[i-1], &alpha)
	}

	// set state
	switch c.rebalanced {
	case false:
		// iy is unused if the database is represented as a vector
		c.state = &itMultiState{
			ix:       index,
			alpha:    alpha,
			a:        a,
			dbLength: cst.DBLength,
		}
	case true:
		panic("not yet implemented")
	}

	// additive secret sharing
	aExtended := []field.Element{field.One()}
	aExtended = append(aExtended, a...)

	// create query vectors for all the servers
	vectors := make([][][]field.Element, numServers)
	for k := range vectors {
		vectors[k] = make([][]field.Element, c.state.dbLength)
		for i := range vectors[0] {
			vectors[k][i] = make([]field.Element, 1+blockLength)
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
	eia := make([][]field.Element, c.state.dbLength)
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
			rand, err := field.RandomVector(1+blockLength, c.xof)
			if err != nil {
				panic(err)
			}
			vectors[k][i] = rand
		}

		// we should perform component-wise additive secret sharing
		for b := 0; b < 1+blockLength; b++ {
			sum := field.Zero()
			for k := 0; k < numServers-1; k++ {
				sum.Add(&sum, &vectors[k][i][b])
			}
			vectors[numServers-1][i][b].Set(&sum)
			vectors[numServers-1][i][b].Neg(&vectors[numServers-1][i][b])
			vectors[numServers-1][i][b].Add(&vectors[numServers-1][i][b], &eia[i][b])
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
			sum[i].Add(&sum[i], &answers[s][i])
		}

	}

	//tag := sum[len(sum)-1]
	messages := sum[:len(sum)-1]

	// compute reconstructed tag
	reconstructedTag := field.Zero()
	for i := 0; i < len(messages); i++ {
		var prod field.Element
		prod.Mul(&c.state.a[i], &messages[i])
		reconstructedTag.Add(&reconstructedTag, &prod)
	}

	//if !tag.Equal(&reconstructedTag) {
	//	return nil, errors.New("REJECT")
	//}

	return messages, nil
}
