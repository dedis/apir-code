package client

import (
	"io"
	"log"
	"math/bits"

	"github.com/dimakogan/dpf-go/dpf"
	"github.com/si-co/vpir-code/lib/database"
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
	return [][]byte{[]byte(queries[0]), []byte(queries[1])}, nil
}

// Query outputs the queries, i.e. DPF keys, for index i. The DPF
// implementation assumes two servers.
func (c *PIRdpf) Query(index, numServers int) []dpf.DPFkey {
	if invalidQueryInputsDPF(index, numServers) {
		log.Fatal("invalid query inputs")
	}
	// set the client state. The entries specific to VPIR are not used
	c.state = &state{
		ix: index / c.dbInfo.NumColumns,
		iy: index % c.dbInfo.NumColumns,
	}

	// compute dpf keys
	key0, key1 := dpf.Gen(uint64(c.state.iy), uint64(bits.Len(uint(c.dbInfo.NumColumns)-1)))

	return []dpf.DPFkey{key0, key1}
}

// ReconstructBytes returns []byte
func (c *PIRdpf) ReconstructBytes(a [][]byte) (interface{}, error) {
	return c.Reconstruct(a)
}

// Reconstruct reconstruct the entry of the database from answers
func (c *PIRdpf) Reconstruct(answers [][]byte) ([]byte, error) {
	return reconstructPIR(answers, c.dbInfo, c.state)
}
