package client

import (
	"io"
	"log"
	"math/bits"

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
	panic("not yet implemented")
	return nil, nil
}

// Query ...
func (c *PIRdpf) Query(index, numServers int) []dpf.DPFkey {
	if invalidQueryInputsDPF(index, numServers) {
		log.Fatal("invalid query inputs")
	}
	var err error
	c.state, err = generateClientState(index, c.rnd, c.dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	// client initialization is the same for both single- and multi-bit scheme
	key0, key1 := dpf.Gen(uint64(c.state.iy), uint64(bits.Len(uint(c.dbInfo.NumColumns))))

	return []dpf.DPFkey{key0, key1}
}

// ReconstructBytes..
func (c *PIRdpf) ReconstructBytes(a [][]byte) ([]field.Element, error) {
	panic("not yet implemented")
	return nil, nil
}

// Reconstruct...
func (c *PIRdpf) Reconstruct(answers [][]byte) ([]byte, error) {
	sum := make([][]byte, c.dbInfo.NumRows)

	// sum answers as vectors in GF(2)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		sum[i] = make([]byte, c.dbInfo.BlockSize)
		for b := 0; b < c.dbInfo.BlockSize; b++ {
			for k := range answers {
				sum[i][b] ^= answers[k][i*c.dbInfo.BlockSize+b]
			}
		}
	}

	return sum[c.state.ix], nil
}
