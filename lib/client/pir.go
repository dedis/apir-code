package client

import (
	"io"
	"log"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic classical PIR client for scheme working in GF(2).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this client, via a boolean variable

// Client for the information theoretic classical PIR single-bit scheme
type PIR struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewItSingleByte return a client for the classical PIR single-bit scheme in
// GF(2), working both with the vector and the rebalanced representation of the
// database.
func NewPIR(rnd io.Reader, info database.Info) *PIR {
	return &PIR{
		rnd:    rnd,
		dbInfo: &info,
		state:  nil,
	}
}

// QueryBytes is wrapper around Query to implement the Client interface
func (c *PIR) QueryBytes(index, numServers int) ([][]byte, error) {
	return c.Query(index, numServers), nil
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *PIR) Query(index int, numServers int) [][]byte {
	if invalidQueryInputsIT(index, numServers) {
		log.Fatal("invalid query inputs")
	}

	// set the client state. The entries specific to VPIR are not used
	c.state = &state{
		ix: index / c.dbInfo.NumColumns,
		iy: index % c.dbInfo.NumColumns,
	}

	vectors, err := c.secretShare(numServers)
	if err != nil {
		log.Fatal(err)
	}

	return vectors
}

func (c *PIR) ReconstructBytes(a [][]byte) ([]field.Element, error) {
	panic("not yet implemented")
	return nil, nil
}

func (c *PIR) Reconstruct(answers [][]byte) ([]byte, error) {
	if c.dbInfo.BlockSize == cst.SingleBitBlockLength {
		// sum answers as vectors in GF(2) only for the
		// row of interest
		sum := make([]byte, c.dbInfo.NumRows)
		for i := 0; i < c.dbInfo.NumRows; i++ {
			for k := range answers {
				sum[i] ^= answers[k][i]
			}
		}

		return []byte{sum[c.state.ix]}, nil
	}
	sum := make([][]byte, c.dbInfo.NumRows)

	// sum answers as vectors in F^(b+1)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		sum[i] = make([]byte, c.dbInfo.BlockSize+1)
		for b := 0; b < c.dbInfo.BlockSize+1; b++ {
			for k := range answers {
				sum[i][b] ^= answers[k][i*(c.dbInfo.BlockSize+1)+b]
			}
		}
	}
	messages := make([]byte, c.dbInfo.BlockSize)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		copy(messages, sum[i][:len(sum[i])-1])
	}

	return sum[c.state.ix][:len(sum[c.state.ix])-1], nil
}

func (c *PIR) secretShare(numServers int) ([][]byte, error) {
	// number of bytes in the whole vector
	blockLen := c.dbInfo.BlockSize
	// handle single bit case
	if blockLen == 0 {
		blockLen = 1
	}
	vectorLen := c.dbInfo.NumColumns * blockLen

	// create query vectors for all the servers F^(1+b)
	vectors := make([][]byte, numServers)
	for k := range vectors {
		vectors[k] = make([]byte, vectorLen)
	}

	// Get random elements for all numServers-1 vectors
	rand := make([]byte, (numServers-1)*vectorLen)
	if _, err := c.rnd.Read(rand); err != nil {
		return nil, err
	}

	// perform additive secret sharing
	for j := 0; j < c.dbInfo.NumColumns; j++ {
		colStart := j * blockLen
		colEnd := (j + 1) * blockLen

		// assign k - 1 random vectors of length dbLength
		for k := 0; k < numServers-1; k++ {
			copy(vectors[k][colStart:colEnd], rand[k*vectorLen+colStart:k*vectorLen+colEnd])
		}

		// we should perform component-wise additive secret sharing
		for b := colStart; b < colEnd; b++ {
			sum := byte(0)
			for k := 0; k < numServers-1; k++ {
				sum ^= vectors[k][b]
			}
			vectors[numServers-1][b] = sum
			// set alpha vector at the block we want to retrieve
			if j == c.state.iy {
				vectors[numServers-1][b] ^= byte(1)
			}
		}
	}

	return vectors, nil
}
