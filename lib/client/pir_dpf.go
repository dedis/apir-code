package client

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"math/bits"

	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"

	"github.com/dimakogan/dpf-go/dpf"
)

// PIRdpf represent the client for the DPF-based multi-bit classical PIR scheme
type PIRdpf struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewPIRdpf returns a new client for the DPF-base multi-bit classical PIR
// scheme
func NewPIRdpf(rnd io.Reader, info *database.Info) *PIRdpf {
	return &PIRdpf{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *PIRdpf) QueryBytes(index, numServers int) ([][]byte, error) {
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

// Query outputs numServers query for index i
func (c *PIRdpf) Query(index, numServers int) []dpf.DPFkey {
	if invalidQueryInputsDPF(index, numServers) {
		log.Fatal("invalid query inputs")
	}
	// set the client state. The entries specific to VPIR are not used
	c.state = &state{
		ix: index / c.dbInfo.NumColumns,
		iy: index % c.dbInfo.NumColumns,
	}
	key0, key1 := dpf.Gen(uint64(c.state.iy), uint64(bits.Len(uint(c.dbInfo.NumColumns)-1)))

	return []dpf.DPFkey{key0, key1}
}

// ReconstructBytes will never be implemented for PIR
func (c *PIRdpf) ReconstructBytes(a [][]byte) ([]field.Element, error) {
	panic("not yet implemented")
	return nil, nil
}

// Reconstruct reconstruct the entry of the database from answers
func (c *PIRdpf) Reconstruct(answers [][]byte) ([]byte, error) {
	sum := make([][]byte, c.dbInfo.NumRows)

	// sum answers as vectors in GF(2)
	bs := c.dbInfo.BlockSize
	for i := 0; i < c.dbInfo.NumRows; i++ {
		sum[i] = make([]byte, c.dbInfo.BlockSize)
		for k := range answers {
			fastxor.Bytes(sum[i], sum[i], answers[k][i*bs:bs*(i+1)])
		}
	}

	return sum[c.state.ix], nil
}
