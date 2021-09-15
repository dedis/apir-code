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
		Fss: fss.ClientInitialize(64, info.BlockSize*field.ConcurrentExecutions), // TODO: solve +1 here, only for VPIR
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *FSS) QueryBytes(in []byte, numServers int) ([][]byte, error) {
	// decode the input query previously encoded as bytes
	q, err := query.DecodeClientFSS(in)
	if err != nil {
		return nil, err
	}

	queries := c.Query(q, numServers)

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
func (c *FSS) Query(q *query.ClientFSS, numServers int) []*query.FSS {
	if invalidQueryInputsFSS(numServers) {
		log.Fatal("invalid query inputs")
	}
	// initialize empty client state
	c.state = &state{}
	// crete state for retrieving a single key, i.e. exact match
	if q.Target == query.Key {
		panic("not yet implemented")
	} else {
		c.state.alphas = make([]uint32, field.ConcurrentExecutions)
		c.state.a = make([]uint32, field.ConcurrentExecutions*2)
		for i := 0; i < field.ConcurrentExecutions; i++ {
			c.state.alphas[i] = field.RandElementWithPRG(c.rnd)
			// c.state.a contains [1, alpha_i] for i = 0, 1, 2, 3
			copy(c.state.a[i*2:i*2+2], []uint32{1, c.state.alphas[i]})
		}
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
	// we keep only the last value of out: if all the answers accepts, then
	// the out value is the same for all of them
	out := uint32(0)
	for i := 0; i < field.ConcurrentExecutions; i++ {
		out = (answers[0][i*2] + answers[1][i*2]) % field.ModP
		tmp := (uint64(out) * uint64(c.state.alphas[i])) % uint64(field.ModP)
		tag := uint32(tmp)
		reconstructedTag := (answers[0][i*2+1] + answers[1][i*2+1]) % field.ModP
		if tag != reconstructedTag {
			return nil, errors.New("REJECT")
		}
	}

	return []uint32{out}, nil
}
