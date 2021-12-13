package client

import (
	"errors"
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
	// AVG case
	if len(answers[0]) == 2*(1+field.ConcurrentExecutions) {
		countFirst := answers[0][:1+field.ConcurrentExecutions]
		countSecond := answers[1][:1+field.ConcurrentExecutions]
		sumFirst := answers[0][1+field.ConcurrentExecutions:]
		sumSecond := answers[1][1+field.ConcurrentExecutions:]

		dataCount := (countFirst[0] + countSecond[0]) % field.ModP
		dataCountCasted := uint64(dataCount)
		sumCount := (sumFirst[0] + sumSecond[0]) % field.ModP
		sumCountCasted := uint64(sumCount)

		// check tags
		for i := 0; i < field.ConcurrentExecutions; i++ {
			tmpCount := (dataCountCasted * uint64(c.state.alphas[i])) % uint64(field.ModP)
			tagCount := uint32(tmpCount)
			reconstructedTagCount := (countFirst[i+1] + countSecond[i+1]) % field.ModP
			if tagCount != reconstructedTagCount {
				return 0, errors.New("REJECT count")
			}

			tmpSum := (sumCountCasted * uint64(c.state.alphas[i])) % uint64(field.ModP)
			tagSum := uint32(tmpSum)
			reconstructedTagSum := (sumFirst[i+1] + sumSecond[i+1]) % field.ModP
			if tagSum != reconstructedTagSum {
				return 0, errors.New("REJECT sum")
			}
		}

		return sumCount / dataCount, nil

	} else {
		// compute data
		data := (answers[0][0] + answers[1][0]) % field.ModP
		dataCasted := uint64(data)

		// check tags
		for i := 0; i < field.ConcurrentExecutions; i++ {
			tmp := (dataCasted * uint64(c.state.alphas[i])) % uint64(field.ModP)
			tag := uint32(tmp)
			reconstructedTag := (answers[0][i+1] + answers[1][i+1]) % field.ModP
			if tag != reconstructedTag {
				return 0, errors.New("REJECT")
			}
		}

		return data, nil
	}
}
