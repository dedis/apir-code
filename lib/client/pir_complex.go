package client

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

// PIRfss represent the client for the FSS-based complex-queries non-verifiable PIR
type PIRfss struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state

	Fss *fss.Fss
}

// NewPIRfss returns a new client for the DPF-base multi-bit classical PIR
// scheme
func NewPIRfss(rnd io.Reader, info *database.Info) *PIRfss {
	return &PIRfss{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
		Fss:    fss.ClientInitialize(),
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
func (c *PIRfss) Query(q *query.ClientFSS, numServers int) []*query.FSS {
	if invalidQueryInputsFSS(numServers) {
		log.Fatal("invalid query inputs")
	}

	fssKeys := c.Fss.GenerateTreePF(q.Input)

	return []*query.FSS{
		{Info: q.Info, FssKey: fssKeys[0]},
		{Info: q.Info, FssKey: fssKeys[1]},
	}
}

// ReconstructBytes returns []byte
// func (c *PIRfss) ReconstructBytes(a [][]byte) (interface{}, error) {
// 	return c.Reconstruct(a)
// }

// Reconstruct reconstruct the entry of the database from answers
func (c *PIRfss) Reconstruct(answers []int) (int, error) {
	return answers[0] + answers[1], nil
}
