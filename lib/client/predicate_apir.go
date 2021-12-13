package client

import (
	"io"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

const BlockLength = 2

// PredicateAPIR represent the client for the FSS-based complex-queries non-verifiable PIR
type PredicateAPIR struct {
	*clientFSS
}

// NewFSS returns a new client for the FSS-based single- and multi-bit schemes
func NewPredicateAPIR(rnd io.Reader, info *database.Info) *PredicateAPIR {
	executions := 1 + field.ConcurrentExecutions
	return &PredicateAPIR{
		&clientFSS{
			rnd:    rnd,
			dbInfo: info,
			state:  nil,
			// one value for the data, four values for the info-theoretic MAC
			Fss:        fss.ClientInitialize(executions),
			executions: executions,
		},
	}
}

// QueryBytes executes Query and encodes the result a byte array for each
// server
func (c *PredicateAPIR) QueryBytes(in []byte, numServers int) ([][]byte, error) {
	return c.queryBytes(in, numServers)
}

// Query takes as input the index of the entry to be retrieved and the number
// of servers (= 2 in the DPF case). It returns the two FSS keys.
func (c *PredicateAPIR) Query(q *query.ClientFSS, numServers int) []*query.FSS {
	return c.query(q, numServers)
}

// ReconstructBytes decodes the answers from the servers and reconstruct the
// entry, returned as []uint32
func (c *PredicateAPIR) ReconstructBytes(a [][]byte) (interface{}, error) {
	answer, err := decodeAnswer(a)
	if err != nil {
		return nil, err
	}

	return c.Reconstruct(answer)
}

// Reconstruct takes as input the answers from the client and returns the
// reconstructed entry after the appropriate integrity check.
func (c *PredicateAPIR) Reconstruct(answers [][]uint32) (uint32, error) {
	return c.reconstruct(answers)
}
