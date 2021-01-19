package client

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"log"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic client for single-bit and multi-bit schemes
// working in F(2^127-1). Both vector and matrix (rebalanced)
// representations of the database are handled by this client, via
// a boolean variable

// ITClient represents the client for the information theoretic multi-bit scheme
type ITClient struct {
	rnd    io.Reader
	dbInfo database.Info
	state  *state
}

// NewITClient returns a client for the information theoretic multi-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITClient(rnd io.Reader, info database.Info) *ITClient {
	return &ITClient{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

func (c *ITClient) QueryBytes(qi []byte) ([]byte, error) {
	queryInputs, err := decodeQueryInputs(qi)
	if err != nil {
		return nil, err
	}

	// get reconstruction
	q := c.Query(queryInputs.index, queryInputs.numServers)

	// encode reconstruction
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(q); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITClient) Query(index, numServers int) [][][]field.Element {
	if invalidQueryInputsIT(index, numServers) {
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

	// Compute the position in the db (vector or matrix)
	ix := index % c.dbInfo.NumColumns
	// if db is a vector, iy always equals 0
	iy := index / c.dbInfo.NumColumns
	// set state
	c.state = &state{
		ix:    ix,
		iy:    iy,
		alpha: alpha,
	}

	if c.dbInfo.BlockSize != cst.SingleBitBlockLength {
		// compute vector a = (1, alpha, alpha^2, ..., alpha^b) for the
		// multi-bit scheme
		// Temporarily +1 to BlockSize for recovering true value
		a = make([]field.Element, c.dbInfo.BlockSize+1)
		a[0] = field.One()
		a[1] = alpha
		for i := 2; i < len(a); i++ {
			a[i].Mul(&a[i-1], &alpha)
		}
		c.state.a = a[1:]
	} else {
		// the single-bit scheme needs a single alpha
		a = make([]field.Element, 1)
		a[0] = alpha
		c.state.a = a
	}

	vectors, err = c.secretShare(a, numServers)
	if err != nil {
		log.Fatal(err)
	}

	return vectors
}

func (c *ITClient) ReconstructBytes(a []byte) ([]byte, error) {
	answer, err := decodeAnswer(a)
	if err != nil {
		return nil, err
	}
	r, err := c.Reconstruct(answer)
	if err != nil {
		return nil, err
	}
	return encodeReconstruct(r)
}

func (c *ITClient) Reconstruct(answers [][][]field.Element) ([]field.Element, error) {
	sum := make([][]field.Element, c.dbInfo.NumRows)

	if c.dbInfo.BlockSize == cst.SingleBitBlockLength {
		// sum answers as vectors in F^b
		for i := 0; i < c.dbInfo.NumRows; i++ {
			sum[i] = make([]field.Element, 1)
			for k := range answers {
				sum[i][0].Add(&sum[i][0], &answers[k][i][0])
			}
		}
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

	// sum answers as vectors in F^(b+1)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		sum[i] = make([]field.Element, c.dbInfo.BlockSize+1)
		for b := 0; b < c.dbInfo.BlockSize+1; b++ {
			for k := range answers {
				sum[i][b].Add(&sum[i][b], &answers[k][i][b])
			}
		}
	}
	var tag, prod field.Element
	messages := make([]field.Element, c.dbInfo.BlockSize)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		copy(messages, sum[i][:len(sum[i])-1])
		tag = sum[i][len(sum[i])-1]
		// compute reconstructed tag
		reconstructedTag := field.Zero()
		for b := 0; b < len(messages); b++ {
			prod.Mul(&c.state.a[b], &messages[b])
			reconstructedTag.Add(&reconstructedTag, &prod)
		}
		if !tag.Equal(&reconstructedTag) {
			return nil, errors.New("REJECT")
		}
	}

	return sum[c.state.iy][:len(sum[c.state.iy])-1], nil
}

// secretShare the vector a among numServers non-colluding servers
func (c *ITClient) secretShare(a []field.Element, numServers int) ([][][]field.Element, error) {
	// get block length
	alen := len(a)

	// create query vectors for all the servers F^(1+b)
	vectors := make([][][]field.Element, numServers)
	for k := range vectors {
		vectors[k] = make([][]field.Element, c.dbInfo.NumColumns)
		for j := 0; j < c.dbInfo.NumColumns; j++ {
			vectors[k][j] = make([]field.Element, alen)
		}
	}

	// Get random elements for all numServers-1 vectors
	rand, err := field.RandomVectors(c.rnd, c.dbInfo.NumColumns*(numServers-1), alen)
	if err != nil {
		return nil, err
	}
	// perform additive secret sharing
	eia := make([][]field.Element, c.dbInfo.NumColumns)
	for j := 0; j < c.dbInfo.NumColumns; j++ {
		// create basic zero vector in F^(1+b)
		eia[j] = field.ZeroVector(alen)

		// set alpha at the index we want to retrieve
		if j == c.state.ix {
			copy(eia[j], a)
		}

		// Assign k - 1 random vectors of length dbLength containing
		// elements in F^(1+b)
		for k := 0; k < numServers-1; k++ {
			vectors[k][j] = rand[k*c.dbInfo.NumColumns+j]
		}

		// we should perform component-wise additive secret sharing
		for b := 0; b < alen; b++ {
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
