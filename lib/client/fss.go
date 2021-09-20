package client

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

const BlockLength = 2

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
		// TODO: avoid hardcoded 64
		Fss: fss.ClientInitialize(64, 1+field.ConcurrentExecutions),
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *FSS) QueryBytes(in []byte, numServers int) ([][]byte, error) {
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

// Query takes as input the index of the entry to be retrieved and the number
// of servers (= 2 in the DPF case). It returns the two FSS keys.
func (c *FSS) Query(q *query.ClientFSS, numServers int) []*query.FSS {
	if invalidQueryInputsFSS(numServers) {
		log.Fatal("invalid query inputs")
	}

	// set client state
	c.state = &state{}
	c.state.alphas = make([]uint32, field.ConcurrentExecutions)
	c.state.a = make([]uint32, field.ConcurrentExecutions+1)
	c.state.a[0] = 1
	for i := 0; i < field.ConcurrentExecutions; i++ {
		c.state.alphas[i] = field.RandElementWithPRG(c.rnd)
		// c.state.a contains [1, alpha_i] for i = 0, .., 3
		c.state.a[i+1] = c.state.alphas[i]
	}

	// client initialization is the same for both single- and multi-bit scheme
	fssKeys := c.Fss.GenerateTreePF(q.Input, c.state.a)

	return []*query.FSS{
		{Target: q.Target, FromStart: q.FromStart, FromEnd: q.FromEnd, FssKey: fssKeys[0]},
		{Target: q.Target, FromStart: q.FromStart, FromEnd: q.FromEnd, FssKey: fssKeys[1]},
	}
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
	// compute data
	data := (answers[0][0] + answers[1][0]) % field.ModP
	dataCasted := uint64(data)

	// check tags
	for i := 0; i < field.ConcurrentExecutions; i++ {
		tmp := (dataCasted * uint64(c.state.alphas[i])) % uint64(field.ModP)
		tag := uint32(tmp)
		reconstructedTag := (answers[0][i+1] + answers[1][i+1]) % field.ModP
		if tag != reconstructedTag {
			return nil, errors.New("REJECT")
		}
	}

	return []uint32{data}, nil
}
