package client

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
)

// FSS represent the client for the FSS-based single- and multi-bit schemes
type FSS struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state

	Fss *fss.Fss
}

// NewFSS returns a new client for the FSS-based single- and multi-bit schemes
func NewFSS(rnd io.Reader, info *database.Info) *FSS {
	return &FSS{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
		Fss:    fss.ClientInitialize(field.Bits, info.BlockSize), // TODO: solve +1 here, only for VPIR
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *FSS) QueryBytes(index, numServers int) ([][]byte, error) {
	queries := c.Query(index, numServers)

	// encode all the queries in bytes
	out := make([][]byte, len(queries))
	for i, q := range queries {
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(q); err != nil {
			return nil, err
		}
		out[i] = buf.Bytes()
	}

	return out, nil
}

// Query takes as input the index of the entry to be retrieved and the number
// of servers (= 2 in the DPF case). It returns the two FSS keys.
func (c *FSS) Query(index, numServers int) []fss.FssKeyEq2P {
	if invalidQueryInputsDPF(index, numServers) {
		log.Fatal("invalid query inputs")
	}
	var err error
	c.state, err = generateClientState(index, c.rnd, c.dbInfo)
	if err != nil {
		log.Fatal(err)
	}

	// client initialization is the same for both single- and multi-bit scheme
	return c.Fss.GenerateTreePF(uint32(index), c.state.a)
}

// ReconstructBytes decodes the answers from the servers and reconstruct the
// entry, returned as []uint32
func (c *FSS) ReconstructBytes(a [][]byte) (interface{}, error) {
	answer, err := decodeAnswer(a)
	if err != nil {
		return nil, err
	}

	return c.Reconstruct(answer)
}

// Reconstruct takes as input the answers from the client and returns the
// reconstructed entry after the appropriate integrity check.
func (c *FSS) Reconstruct(answers [][]uint32) ([]uint32, error) {
	return reconstruct(answers, c.dbInfo, c.state)
}
