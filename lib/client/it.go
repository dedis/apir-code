package client

import (
	"errors"
	"fmt"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"io"
	"log"
)

// Information theoretic client for single-bit and multi-bit schemes
// working in F(2^127-1). Both vector and matrix (rebalanced)
// representations of the database are handled by this client, via
// a boolean variable

// ITClient represents the client for the information theoretic multi-bit scheme
type ITClient struct {
	rnd    io.Reader
	state  *itState
	dbInfo database.Info
}

type itState struct {
	ix    int
	iy    int // unused if not rebalanced
	alpha field.Element
	a     []field.Element
}

// NewITSingleGF return a client for the information theoretic multi-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITClient(rnd io.Reader, info database.Info) *ITClient {
	return &ITClient{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITClient) Query(index, numServers int) [][][]field.Element {
	if invalidQueryInputs(index, numServers) {
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

	if c.dbInfo.BlockSize != cst.SingleBitBlockLength {
		// compute vector a = (1, alpha, alpha^2, ..., alpha^b) for the
		// multi-bit scheme
		// +1 for recovering true value
		c.dbInfo.BlockSize += 1
		a = make([]field.Element, c.dbInfo.BlockSize)
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
	ix := index % c.dbInfo.NumColumns
	// if db is a vector, iy always equals 0
	iy := index / c.dbInfo.NumColumns
	c.state = &itState{
		ix:       ix,
		iy:       iy,
		alpha:    alpha,
		a:        a[1:],
	}

	vectors, err = c.secretShare(a, numServers)
	if err != nil {
		log.Fatal(err)
	}

	return vectors
}

func (c *ITClient) Reconstruct(answers [][][]field.Element) ([]field.Element, error) {
	fmt.Println(answers[0])
	sum := make([][]field.Element, c.dbInfo.NumRows)
	// sum answers as vectors in F(2^128)^(b+1)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		sum[i] = make([]field.Element, c.dbInfo.BlockSize)
		for b := 0; b < c.dbInfo.BlockSize; b++ {
			for k := range answers {
				sum[i][b].Add(&sum[i][b], &answers[k][i][b])
			}
		}
	}

	if c.dbInfo.BlockSize == cst.SingleBitBlockLength {
		for i := 0; i < c.dbInfo.NumRows; i++ {
			if i == c.state.iy {
				switch {
				case sum[i][0].Equal(&c.state.alpha):
					return []field.Element{cst.One}, nil
				case sum[i][0].Equal(&cst.Zero):
					return []field.Element{cst.Zero}, nil
				default:
					return nil, errors.New("REJECT!")
				}
			} else {
				if !sum[i][0].Equal(&c.state.alpha) && !sum[i][0].Equal(&cst.Zero) {
					return nil, errors.New("REJECT!")
				}
			}
		}
	}

	var tag, prod field.Element
	var messages []field.Element
	for i := 0; i < c.dbInfo.NumRows; i++ {
		tag = sum[i][len(sum)-1]
		messages = sum[i][:len(sum)-1]
		// compute reconstructed tag
		reconstructedTag := field.Zero()
		for i := 0; i < len(messages); i++ {
			prod.Mul(&c.state.a[i], &messages[i])
			reconstructedTag.Add(&reconstructedTag, &prod)
		}
		if !tag.Equal(&reconstructedTag) {
			return nil, errors.New("REJECT")
		}
	}

	return sum[c.state.iy][:len(sum)-1], nil
}

// secretShare the vector a among numServers non-colluding servers
func (c *ITClient) secretShare(a []field.Element, numServers int) ([][][]field.Element, error) {
	// get block length
	blockSize := len(a)

	// create query vectors for all the servers
	vectors := make([][][]field.Element, numServers)
	for k := range vectors {
		vectors[k] = make([][]field.Element, c.dbInfo.NumColumns)
		for j := 0; j < c.dbInfo.NumColumns; j++ {
			vectors[k][j] = make([]field.Element, blockSize)
		}
	}

	// Get random elements for all numServers-1 vectors
	rand, err := field.RandomVectors(c.rnd, c.dbInfo.NumColumns*(numServers-1), blockSize)
	if err != nil {
		return nil, err
	}
	// perform additive secret sharing
	eia := make([][]field.Element, c.dbInfo.NumColumns)
	for j := 0; j < c.dbInfo.NumColumns; j++ {
		// create basic zero vector in F^(b)
		eia[j] = field.ZeroVector(blockSize)

		// set alpha at the index we want to retrieve
		if j == c.state.ix {
			copy(eia[j], a)
		}

		// Assign k - 1 random vectors of length dbLength containing
		// elements in F^(b)
		for k := 0; k < numServers-1; k++ {
			vectors[k][j] = rand[k*c.dbInfo.NumColumns+j]
		}

		// we should perform component-wise additive secret sharing
		for b := 0; b < blockSize; b++ {
			sum := field.Zero()
			for k := 0; k < numServers-1; k++ {
				sum.Add(&sum, &vectors[k][j][b])
			}
			vectors[numServers-1][j][b].Set(&sum)
			vectors[numServers-1][j][b].Neg(&vectors[numServers-1][j][b])
			vectors[numServers-1][j][b].Add(&vectors[numServers-1][j][b], &eia[j][b])
		}
	}

	return vectors, nil
}
