package client

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

type clientFSS struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state

	Fss        *fss.Fss
	executions int
}

func (c *clientFSS) queryBytes(in []byte, numServers int) ([][]byte, error) {
	inQuery, err := query.DecodeClientFSS(in)
	if err != nil {
		return nil, err
	}

	queries := c.query(inQuery, numServers)

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

func (c *clientFSS) query(q *query.ClientFSS, numServers int) []*query.FSS {
	if invalidQueryInputsFSS(numServers) {
		log.Fatal("invalid query inputs")
	}

	// set client state
	c.state = &state{}
	c.state.alphas = make([]uint32, c.executions)
	c.state.a = make([]uint32, c.executions)
	c.state.a[0] = 1 // to retrieve data
	// below executed only for authenticated. The -1 is to ignore
	// the value for the data already initialized
	for i := 0; i < c.executions-1; i++ {
		c.state.alphas[i] = field.RandElementWithPRG(c.rnd)
		// c.state.a contains [1, alpha_i] for i = 0, .., 3
		c.state.a[i+1] = c.state.alphas[i]
	}

	// generate FSS keys
	fssKeys := c.Fss.GenerateTreePF(q.Input, c.state.a)

	return []*query.FSS{
		{Info: q.Info, FssKey: fssKeys[0]},
		{Info: q.Info, FssKey: fssKeys[1]},
	}
}

func (c *clientFSS) generateFSSKeys(q *query.ClientFSS, a []uint32) []*query.FSS {
	fssKeys := c.Fss.GenerateTreePF(q.Input, a)

	return []*query.FSS{
		{Info: q.Info, FssKey: fssKeys[0]},
		{Info: q.Info, FssKey: fssKeys[1]},
	}
}
