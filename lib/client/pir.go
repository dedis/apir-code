package client

import (
	"crypto/rand"
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

	vectors, err := c.secretSharing(numServers)
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
	sum := make([][]byte, c.dbInfo.NumRows)

	if c.dbInfo.BlockSize == cst.SingleBitBlockLength {
		// sum answers as vectors in GF(2) only for the
		// row of interest
		for i := 0; i < c.dbInfo.NumRows; i++ {
			sum[i] = make([]byte, 1)
			for k := range answers {
				sum[i][0] ^= answers[k][i]
			}
		}

		return []byte{sum[c.state.ix][0]}, nil
	}

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

func (c *PIR) secretSharing(numServers int) ([][]byte, error) {
	ei := make([]byte, c.dbInfo.NumColumns)
	ei[c.state.ix] = byte(1)

	vectors := make([][]byte, numServers)

	// create query vectors for all the servers
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]byte, c.dbInfo.NumColumns)
	}

	zero := byte(0)

	// for all except one server, we need dbLength random elements
	// to perform the secret sharing
	b := make([]byte, c.dbInfo.NumColumns*(numServers-1))
	_, err := rand.Read(b)
	if err != nil {
		panic("error in randomness generation")
	}
	for i := 0; i < c.dbInfo.NumColumns; i++ {
		sum := zero
		for k := 0; k < numServers-1; k++ {
			rand := b[c.dbInfo.NumColumns*k+i] % 2
			vectors[k][i] = rand
			sum ^= rand
		}
		vectors[numServers-1][i] = ei[i] ^ sum
	}

	return vectors, nil
}
