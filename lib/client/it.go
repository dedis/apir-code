package client

import (
	"errors"
	"io"
	"log"
	"math"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic client for single-bit and multi-bit schemes
// working in F(2^127-1). Both vector and matrix (rebalanced)
// representations of the database are handled by this client, via
// a boolean variable

// ITMulti represents the client for the information theoretic multi-bit scheme
type ITMulti struct {
	rnd        io.Reader
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
func NewITMulti(rnd io.Reader, rebalanced bool) *ITMulti {
	return &ITMulti{
		rnd:        rnd,
		rebalanced: rebalanced,
		state:      nil,
	}
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITMulti) Query(index, blockSize, numServers int) [][][]field.Element {
	if invalidQueryInputs(index, blockSize, numServers) {
		log.Fatal("invalid query inputs")
	}
	var alpha field.Element
	var a []field.Element
	var vectors [][][]field.Element
	var err error

	// sample random alpha using blake2b
	if _, err = alpha.SetRandom(c.rnd); err != nil {
		log.Fatal(err)
	}

	if blockSize != cst.SingleBitBlockLength {
		// compute vector a = (1, alpha, alpha^2, ..., alpha^b) for the multi-bit scheme
		a = make([]field.Element, blockSize + 1)
		a[0] = field.One()
		a[1] = alpha
		for i := 2; i < len(a); i++ {
			a[i].Mul(&a[i-1], &alpha)
		}
	} else {
		// the single-bit scheme needs a single alpha
		a = make([]field.Element, 1)
		a[0] = alpha
	}

	// set state
	switch c.rebalanced {
	case false:
		// iy is unused if the database is represented as a vector
		c.state = &itMultiState{
			ix:       index,
			alpha:    alpha,
			a:        a[1:],
			dbLength: cst.DBLength,
		}
	case true:
		// verified at server side if integer square
		dbLengthSqrt := int(math.Sqrt(cst.DBLength))
		ix := index % dbLengthSqrt
		iy := index / dbLengthSqrt

		c.state = &itMultiState{
			ix:       ix,
			iy:       iy,
			alpha:    alpha,
			a:        a[1:],
			dbLength: dbLengthSqrt,
		}
	}

	vectors, err = c.secretShare(a, numServers)
	if err != nil {
		log.Fatal(err)
	}

	return vectors
}

func (c *ITMulti) Reconstruct(answers [][]field.Element, blockSize int) ([]field.Element, error) {
	answersLen := len(answers[0])
	sum := make([]field.Element, answersLen)

	// sum answers as vectors in F(2^128)^(1+b)
	for i := 0; i < answersLen; i++ {
		sum[i] = field.Zero()
		for s := range answers {
			sum[i].Add(&sum[i], &answers[s][i])
		}
	}

	// select index depending on the matrix representation
	i := 0
	if c.rebalanced {
		i = c.state.iy
	}

	if blockSize == cst.SingleBitBlockLength {
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

// secretShare the vector a among numServers non-colluding servers
func (c *ITMulti) secretShare(a []field.Element, numServers int) ([][][]field.Element, error) {
	// get block length
	blockSize := len(a)

	// create query vectors for all the servers
	vectors := make([][][]field.Element, numServers)
	for k := range vectors {
		vectors[k] = make([][]field.Element, c.state.dbLength)
		for i := 0; i < c.state.dbLength; i++ {
			vectors[k][i] = make([]field.Element, blockSize)
		}
	}

	// Get random elements for all numServers-1 vectors
	rand, err := field.RandomVectors(c.rnd, c.state.dbLength*(numServers-1), blockSize)
	if err != nil {
		return nil, err
	}
	// perform additive secret sharing
	eia := make([][]field.Element, c.state.dbLength)
	for i := 0; i < c.state.dbLength; i++ {
		// create basic zero vector in F^(b)
		eia[i] = field.ZeroVector(blockSize)

		// set alpha at the index we want to retrieve
		if i == c.state.ix {
			copy(eia[i], a)
		}

		// Assign k - 1 random vectors of length dbLength containing
		// elements in F^(b)
		for k := 0; k < numServers-1; k++ {
			vectors[k][i] = rand[k*c.state.dbLength + i]
		}

		// we should perform component-wise additive secret sharing
		for b := 0; b < blockSize; b++ {
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
