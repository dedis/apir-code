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
// handled by this client.

// Client for the information theoretic classical PIR multi-bit scheme
type PIR struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewItSingleByte return a client for the classical PIR multi-bit scheme in
// GF(2), working both with the vector and the rebalanced representation of the
// database.
func NewPIR(rnd io.Reader, info *database.Info) *PIR {
	if info.BlockSize == cst.SingleBitBlockLength {
		panic("single-bit classical PIR protocol not implemented")
	}
	return &PIR{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

// QueryBytes is wrapper around Query to implement the Client interface
func (c *PIR) QueryBytes(index, numServers int) ([][]byte, error) {
	return c.Query(index, numServers), nil
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the database representation
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

// ReconReconstructBytes will never the implemented for PIR
func (c *PIR) ReconstructBytes(a [][]byte) ([]field.Element, error) {
	panic("not yet implemented")
	return nil, nil
}

// Reconstruct reconstruct the entry of the database from answers
func (c *PIR) Reconstruct(answers [][]byte) ([]byte, error) {
	return reconstructPIR(answers, c.dbInfo, c.state)
}

func (c *PIR) secretShare(numServers int) ([][]byte, error) {
	// length of query vector
	vectorLen := c.dbInfo.NumColumns

	// create query vectors for all the servers
	vectors := make([][]byte, numServers)
	for k := range vectors {
		vectors[k] = make([]byte, vectorLen)
	}

	// Get random elements for all numServers-1 vectors
	// TODO: these are actually too many bits, we need
	// 1/8 of them and extract the random bits from
	// the bytes
	rand := make([]byte, (numServers-1)*vectorLen)
	if _, err := c.rnd.Read(rand); err != nil {
		return nil, err
	}

	// perform secret sharing
	for j := 0; j < c.dbInfo.NumColumns; j++ {
		// assign k - 1 random bits for this column
		sum := byte(0)
		for k := 0; k < numServers-1; k++ {
			vectors[k][j] = rand[j+k] % 2
			sum ^= rand[j+k] % 2
		}

		// vectors[numerServers-1][j] is initialized at zero
		vectors[numServers-1][j] ^= sum
		// set alpha vector at the block we want to retrieve
		if j == c.state.iy {
			// we have to add up 1 in this case, as the value
			// is initialized at zero
			vectors[numServers-1][j] ^= byte(1)
		}
	}

	return vectors, nil
}
