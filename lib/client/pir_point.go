package client

import (
	"encoding/binary"
	"io"
	"log"
	"math/bits"

	"github.com/dkales/dpf-go/dpf"
	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/utils"
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

// NewPIR return a client for the classical PIR multi-bit scheme in
// GF(2), working both with the vector and the rebalanced representation of the
// database.
func NewPIR(rnd io.Reader, info *database.Info) *PIR {
	return &PIR{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

// QueryBytes is wrapper around Query to implement the Client interface
func (c *PIR) QueryBytes(in []byte, numServers int) ([][]byte, error) {
	index := int(binary.BigEndian.Uint32(in))
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
	ix, iy := utils.VectorToMatrixIndices(index, c.dbInfo.NumColumns)
	c.state = &state{
		ix: ix,
		iy: iy,
	}

	if numServers == 2 {
		k0, k1 := dpf.Gen(uint64(c.state.iy), uint64(bits.Len(uint(c.dbInfo.NumColumns))))
		return [][]byte{k0, k1}
	}

	vectors, err := c.secretShare(numServers)
	if err != nil {
		log.Fatal(err)
	}

	return vectors
}

// ReconstructBytes returns []byte
func (c *PIR) ReconstructBytes(a [][]byte) (interface{}, error) {
	return c.Reconstruct(a)
}

// Reconstruct reconstruct the entry of the database from answers
func (c *PIR) Reconstruct(answers [][]byte) ([]byte, error) {
	return reconstructPIR(answers, c.dbInfo, c.state)
}

func (c *PIR) secretShare(numServers int) ([][]byte, error) {
	// length of query vector
	// one query bit per column
	vectorLen := c.dbInfo.NumColumns/8 + 1

	// create query vectors for all the servers
	vectors := make([][]byte, numServers)
	for k := range vectors {
		vectors[k] = make([]byte, vectorLen)
	}

	// Get random elements for all numServers-1 vectors.
	// This is faster than extracting single bits
	rand := make([]byte, (numServers-1)*vectorLen)
	if _, err := c.rnd.Read(rand); err != nil {
		return nil, err
	}

	// perform secret sharing
	// find the byte corresponding to the retrieval bit
	// what value this byte should get
	index := c.state.iy / 8
	value := 1 << (c.state.iy % 8)
	vectors[numServers-1][index] = byte(value)
	for k := 0; k < numServers-1; k++ {
		copy(vectors[k], rand[k*vectorLen:(k+1)*vectorLen])
		fastxor.Bytes(vectors[numServers-1], vectors[numServers-1], vectors[k])
	}

	return vectors, nil
}
