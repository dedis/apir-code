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
		Fss: fss.ClientInitialize(64, info.BlockSize), // TODO: solve +1 here, only for VPIR
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
// TODO: this should be changed, how should we manage the new interface?
// Should we add the query typo to the FSS client, or drop the idea of interface
// since now we should have less schemes? To discuss
func (c *FSS) QueryBytes(index, numServers int) ([][]byte, error) {
	// TODO: fix this, here query.Type is hardcoded
	// FIX THIS
	queries := c.Query(index, []bool{false}, query.KeyId, numServers)

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
func (c *FSS) Query(index int, q []bool, t query.Target, numServers int) []*query.FSS {
	if invalidQueryInputsDPF(index, numServers) {
		log.Fatal("invalid query inputs")
	}
	var err error
	c.state, err = generateClientState(index, c.rnd, c.dbInfo)
	c.state.a = make([]uint32, 2)
	c.state.a[0] = 1
	c.state.a[1] = c.state.alpha
	if err != nil {
		log.Fatal(err)
	}

	// client initialization is the same for both single- and multi-bit scheme
	fssKeys := c.Fss.GenerateTreePF(q, c.state.a)

	return []*query.FSS{
		{Target: t, FssKey: fssKeys[0]},
		{Target: t, FssKey: fssKeys[1]},
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
	out := (answers[0][0] + answers[1][0]) % field.ModP
	tmp := (uint64(out) * uint64(c.state.alpha)) % uint64(field.ModP)
	tag := uint32(tmp)
	reconstructedTag := (answers[0][1] + answers[1][1]) % field.ModP
	if tag == reconstructedTag {
		count[0] = out
		return []uint32{out}, nil
	}

	return nil, errors.New("REJECT")
	//return reconstruct(answers, c.dbInfo, c.state)
}
