package client

import (
	"errors"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"golang.org/x/crypto/blake2b"
	"log"
)

// Information theoretic client for single-bit and multi-bit schemes
// working in F(2^127-1). Both vector and matrix (rebalanced)
// representations of the database are handled by this client, via
// a boolean variable

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
func (c *ITMulti) Query(index, blockSize, numServers int) [][][]field.Element {
	var alpha field.Element
	var vectors [][][]field.Element
	var err error

	if invalidQueryInputs(index, blockSize, numServers) {
		log.Fatal("invalid query inputs")
	}

	// sample random alpha using blake2b
	alpha.SetRandom(c.xof)

	// compute vector a = (alpha, alpha^2, ..., alpha^b)
	a := make([]field.Element, blockSize)
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

	if blockSize > 1 {
		// Inserting secret-shared 1 into vectors before alphas
		//a = append(a, field.Zero())
		//copy(a[1:], a)
		//a[0] = field.One()
		a = append([]field.Element{field.One()}, a...)
		vectors, err = c.secretShare(a, blockSize + 1, numServers)
	} else {
		vectors, err = c.secretShare(a, blockSize, numServers)
	}
	if err != nil {
		log.Fatal(err)
	}

	return vectors
}

func (c *ITMulti) Reconstruct(answers [][]field.Element, blockSize int) ([]field.Element, error) {
	answersLen := len(answers[0])
	sum := make([]field.Element, answersLen)

	// sum answers as vectors in GF(2^128)^(1+b)
	for i := 0; i < answersLen; i++ {
		sum[i] = field.Zero()
		for s := range answers {
			sum[i].Add(&sum[i], &answers[s][i])
		}

	}

	i := 0
	if blockSize == 1 {
		switch {
		case sum[i].Equal(&c.state.alpha):
			return []field.Element{cst.One}, nil
		case sum[i].Equal(&cst.Zero):
			return []field.Element{cst.Zero}, nil
		default:
			return nil, errors.New("REJECT!")
		}
	}

	tag := sum[len(sum)-1]
	messages := sum[:len(sum)-1]

	// compute reconstructed tag
	reconstructedTag := field.Zero()
	for i := 0; i < len(messages); i++ {
		var prod field.Element
		prod.Mul(&c.state.a[i], &messages[i])
		reconstructedTag.Add(&reconstructedTag, &prod)
	}

	if !tag.Equal(&reconstructedTag) {
		return nil, errors.New("REJECT")
	}

	return messages, nil
}

// Additive secret sharing
func (c *ITMulti) secretShare(a []field.Element, size, numServers int) ([][][]field.Element, error){
	// create query vectors for all the servers
	vectors := make([][][]field.Element, numServers)
	for k := range vectors {
		vectors[k] = make([][]field.Element, c.state.dbLength)
		for i := 0; i < c.state.dbLength; i++ {
			vectors[k][i] = make([]field.Element, size)
		}
	}

	// perform additive secret sharing
	eia := make([][]field.Element, c.state.dbLength)
	for i := 0; i < c.state.dbLength; i++ {
		// create basic zero vector in F^(b)
		eia[i] = field.ZeroVector(size)

		// set alpha at the index we want to retrieve
		if i == c.state.ix {
			copy(eia[i], a)
		}

		// create k - 1 random vectors of length dbLength containing
		// elements in F^(b)
		for k := 0; k < numServers-1; k++ {
			rand, err := field.RandomVector(size, c.xof)
			if err != nil {
				return nil, err
			}
			vectors[k][i] = rand
		}

		// we should perform component-wise additive secret sharing
		for b := 0; b < size; b++ {
			sum := field.Zero()
			for k := 0; k < numServers-1; k++ {
				sum.Add(&sum, &vectors[k][i][b])
			}
			vectors[numServers-1][i][b].Set(&sum)
			vectors[numServers-1][i][b].Neg(&vectors[numServers-1][i][b])
			vectors[numServers-1][i][b].Add(&vectors[numServers-1][i][b], &eia[i][b])
		}
	}

	return vectors, nil
}
