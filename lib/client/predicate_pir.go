package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

// PredicatePIR represent the client for the FSS-based complex-queries non-verifiable PIR
type PredicatePIR struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state

	Fss *fss.Fss
}

// NewPredicatePIR returns a new client for the DPF-base multi-bit classical PIR
// scheme
func NewPredicatePIR(rnd io.Reader, info *database.Info) *PredicatePIR {
	return &PredicatePIR{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
		Fss:    fss.ClientInitialize(1), // only one value
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *PredicatePIR) QueryBytes(in []byte, numServers int) ([][]byte, error) {
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
func (c *PredicatePIR) Query(q *query.ClientFSS, numServers int) []*query.FSS {
	if invalidQueryInputsFSS(numServers) {
		log.Fatal("invalid query inputs")
	}

	fssKeys := c.Fss.GenerateTreePF(q.Input, []uint32{1})

	return []*query.FSS{
		{Info: q.Info, FssKey: fssKeys[0]},
		{Info: q.Info, FssKey: fssKeys[1]},
	}
}

// ReconstructBytes returns []byte
func (c *PredicatePIR) ReconstructBytes(answers [][]byte) (interface{}, error) {
	in := make([]uint32, 2)
	for i, a := range answers {
		in[i] = binary.BigEndian.Uint32(a)
	}

	return c.Reconstruct(in), nil
}

// Reconstruct reconstruct the entry of the database from answers
func (c *PredicatePIR) Reconstruct(answers []uint32) uint32 {
	return (answers[0] + answers[1]) % field.ModP
}
