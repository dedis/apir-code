package client

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"

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
	dbInfo *database.Info
	state  *state
}

// NewITClient returns a client for the information theoretic multi-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITClient(rnd io.Reader, info database.Info) *ITClient {
	return &ITClient{
		rnd:    rnd,
		dbInfo: &info,
		state:  nil,
	}
}

func (c *ITClient) QueryBytes(index, numServers int) ([][]byte, error) {
	// get reconstruction
	queries := c.Query(index, numServers)

	// encode all the queries in bytes
	out := make([][]byte, len(queries))
	var buf bytes.Buffer
	for i, q := range queries {
		buf.Reset()
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(q); err != nil {
			return nil, err
		}
		out[i] = buf.Bytes()
	}

	return out, nil
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITClient) Query(index, numServers int) [][][]field.Element {
	if invalidQueryInputsIT(index, numServers) {
		log.Fatal("invalid query inputs")
	}
	var err error
	c.state, err = generateClientState(index, c.rnd, c.dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	vectors, err := c.secretShare(numServers)
	if err != nil {
		log.Fatal(err)
	}
	return vectors
}

func (c *ITClient) ReconstructBytes(a [][]byte) ([]field.Element, error) {
	answer, err := decodeAnswer(a)
	if err != nil {
		return nil, err
	}

	return c.Reconstruct(answer)
}

func (c *ITClient) Reconstruct(answers [][][]field.Element) ([]field.Element, error) {
	return reconstruct(answers, c.dbInfo, c.state)
}

// secretShare the vector a among numServers non-colluding servers
func (c *ITClient) secretShare(numServers int) ([][][]field.Element, error) {
	// get block length
	alen := len(c.state.a)

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
		if j == c.state.iy {
			copy(eia[j], c.state.a)
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
