package client

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/bits"

	"github.com/dimakogan/dpf-go/dpf"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	merkletree "github.com/wealdtech/go-merkletree"
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
	if info.BlockSize == cst.SingleBitBlockLength {
		panic("single-bit classical PIR protocol not implemented")
	}
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
	switch c.dbInfo.PIRType {
	case "classical", "":
		return reconstructPIR(answers, c.dbInfo, c.state)
	case "merkle":
		block, err := reconstructPIR(answers, c.dbInfo, c.state)
		if err != nil {
			return block, err
		}
		data := block[:c.dbInfo.BlockSize-c.dbInfo.ProofLen]

		// check Merkle proof
		encodedProof := block[c.dbInfo.BlockSize-c.dbInfo.ProofLen:]
		fmt.Println("encoded proof:", encodedProof)
		proof := database.DecodeProof(encodedProof)
		verified, err := merkletree.VerifyProof(data, proof, c.dbInfo.Root)
		if err != nil {
			log.Fatalf("impossible to verify proof: %v", err)
		}
		fmt.Println(verified)

		return data, nil
	default:
		panic("unknow PIRType")

	}
}
