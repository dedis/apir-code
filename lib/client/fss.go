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

func (c *clientFSS) reconstructBytes(answers [][]byte) (interface{}, error) {
	answer, err := decodeAnswer(answers)
	if err != nil {
		return nil, err
	}

	return c.reconstruct(answer)
}

func (c *clientFSS) reconstruct(answers [][]uint32) (uint32, error) {
	// AVG case
	if len(answers[0]) == 2*c.executions {
		countFirst := answers[0][:c.executions]
		countSecond := answers[1][:c.executions]
		sumFirst := answers[0][c.executions:]
		sumSecond := answers[1][c.executions:]

		dataCount := (countFirst[0] + countSecond[0]) % field.ModP
		dataCountCasted := uint64(dataCount)
		sumCount := (sumFirst[0] + sumSecond[0]) % field.ModP
		sumCountCasted := uint64(sumCount)

		// check tags, executed only for authenticated. The -1 is to ignore
		// the value for the data already initialized
		for i := 0; i < c.executions-1; i++ {
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

		// check tags, executed only for authenticated. The -1 is to ignore
		// the value for the data already initialized
		for i := 0; i < c.executions-1; i++ {
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
