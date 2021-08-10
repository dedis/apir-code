package client

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/utils"
)

// Information theoretic client for single-bit and multi-bit schemes
// working in F(2^127-1). Both vector and matrix (rebalanced)
// representations of the database are handled by this client, via
// a boolean variable

// IT represents the client for the information theoretic multi-bit scheme
type IT struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewIT returns a client for the information theoretic multi-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewIT(rnd io.Reader, info *database.Info) *IT {
	return &IT{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

func (c *IT) QueryBytes(index, numServers int) ([][]byte, error) {
	// get reconstruction
	queries := c.Query(index, numServers)

	// encode all the queries in bytes
	out := make([][]byte, len(queries))
	for i := range queries {
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, queries[i])
		if err != nil {
			return nil, err
		}
		out[i] = buf.Bytes()
	}

	return out, nil
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *IT) Query(index, numServers int) [][]uint32 {
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

// ReconstructBytes returns []uint32
func (c *IT) ReconstructBytes(a [][]byte) (interface{}, error) {
	res := make([][]uint32, len(a))

	for i := range res {
		buf := bytes.NewReader(a[i])
		res[i] = make([]uint32, len(a)/4)
		err := binary.Read(buf, binary.LittleEndian, &res[i])
		if err != nil {
			return nil, err
		}
	}

	return c.Reconstruct(res)
}

func (c *IT) Reconstruct(answers [][]uint32) ([]uint32, error) {
	return reconstruct(answers, c.dbInfo, c.state)
}

// secretShare the vector a among numServers non-colluding servers
func (c *IT) secretShare(numServers int) ([][]uint32, error) {
	// get block length
	blockLen := len(c.state.a)
	// Number of field elements in the whole vector
	vectorLen := c.dbInfo.NumColumns * blockLen

	// create query vectors for all the servers F^(1+b)
	vectors := make([][]uint32, numServers)
	for k := range vectors {
		vectors[k] = make([]uint32, vectorLen)
	}

	// Get random elements for all numServers-1 vectors
	rand := make([]uint32, (numServers-1)*vectorLen)
	var err error
	for i := range rand {
		rand[i], err = utils.RandUint32()
		if err != nil {
			return nil, err
		}

	}

	// perform additive secret sharing
	var colStart, colEnd int
	for j := 0; j < c.dbInfo.NumColumns; j++ {
		colStart = j * blockLen
		colEnd = (j + 1) * blockLen
		// Assign k - 1 random vectors of length dbLength containing
		// elements in F^(1+b)
		for k := 0; k < numServers-1; k++ {
			copy(vectors[k][colStart:colEnd], rand[k*vectorLen+colStart:k*vectorLen+colEnd])
		}

		// we should perform component-wise additive secret sharing
		for b := colStart; b < colEnd; b++ {
			sum := uint32(0)
			for k := 0; k < numServers-1; k++ {
				sum += vectors[k][b] % constants.ModP
			}
			vectors[numServers-1][b] = sum
			vectors[numServers-1][b] = constants.ModP - vectors[numServers-1][b]
			// set alpha vector at the block we want to retrieve
			if j == c.state.iy {
				vectors[numServers-1][b] += c.state.a[b-j*blockLen] % constants.ModP
			}
		}
	}

	return vectors, nil
}
