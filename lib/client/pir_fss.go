package client

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"math/bits"

	"github.com/dimakogan/dpf-go/dpf"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/query"
)

// PIRfss represent the client for the FSS-based complex-queries non-verifiable PIR
type PIRfss struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewPIRfss returns a new client for the DPF-base multi-bit classical PIR
// scheme
func NewPIRfss(rnd io.Reader, info *database.Info) *PIRfss {
	return &PIRfss{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *PIRfss) QueryBytes(in []byte, numServers int) ([][]byte, error) {
	inQuery, err := query.DecodeClientFSS(in)
	if err != nil {
		return nil, err
	}

	queries := c.Query(inQuery, numServers)

	// encode all the queries in bytes
	data := make([][]byte, len(queries))
	for i, q := range queries {
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(q); err != nil {
			return nil, err
		}
		data[i] = buf.Bytes()
	}

	return data, nil
}

// Query outputs the queries, i.e. DPF keys, for index i. The DPF
// implementation assumes two servers.
func (c *PIRfss) Query(q *query.ClientFSS, numServers int) []dpf.DPFkey {
	if invalidQueryInputsFSS(numServers) {
		log.Fatal("invalid query inputs")
	}

	// compute dpf keys
	key0, key1 := dpf.Gen(uint64(c.state.iy), uint64(bits.Len(uint(c.dbInfo.NumColumns)-1)))

	return []dpf.DPFkey{key0, key1}
}

// ReconstructBytes returns []byte
func (c *PIRfss) ReconstructBytes(a [][]byte) (interface{}, error) {
	return c.Reconstruct(a)
}

// Reconstruct reconstruct the entry of the database from answers
func (c *PIRfss) Reconstruct(answers [][]byte) ([]byte, error) {
	return reconstructPIR(answers, c.dbInfo, c.state)
}
