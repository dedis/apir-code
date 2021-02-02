package client

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"math/bits"

	"github.com/si-co/vpir-code/lib/database"

	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

// DPF represent the client for the DPF-based single- and multi-bit schemes
type DPF struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewDPF returns a new client for the DPF-based single- and multi-bit schemes
func NewDPF(rnd io.Reader, info *database.Info) *DPF {
	return &DPF{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *DPF) QueryBytes(index, numServers int) ([][]byte, error) {
	queries := c.Query(index, numServers)

	// encode all the queries in bytes
	out := make([][]byte, len(queries))
	for i, q := range queries {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(q); err != nil {
			return nil, err
		}
		out[i] = buf.Bytes()
	}

	return out, nil
}

// Query ...
func (c *DPF) Query(index, numServers int) []dpf.DPFkey {
	if invalidQueryInputsDPF(index, numServers) {
		log.Fatal("invalid query inputs")
	}
	var err error
	c.state, err = generateClientState(index, c.rnd, c.dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	// client initialization is the same for both single- and multi-bit scheme
	key0, key1 := dpf.Gen(uint64(c.state.iy), c.state.a, uint64(bits.Len(uint(c.dbInfo.NumColumns))))

	return []dpf.DPFkey{key0, key1}
}

// ReconstructBytes ...
func (c *DPF) ReconstructBytes(a [][]byte) ([]field.Element, error) {
	answer, err := decodeAnswer(a)

	if err != nil {
		return nil, err
	}

	return c.Reconstruct(answer)
}

// Reconstruct ..
func (c *DPF) Reconstruct(answers [][]field.Element) ([]field.Element, error) {
	return reconstruct(answers, c.dbInfo, c.state)
}
