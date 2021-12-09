package server

import (
	"bytes"
	"encoding/gob"
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
)

// PredicateAPIR represent the server for the FSS-based complex-queries authenticated PIR
type PredicateAPIR struct {
	ServerFSS
}

func NewPredicateAPIR(db *database.DB, serverNum byte, cores ...int) *PredicateAPIR {
	// use variadic argument for cores to achieve backward compatibility
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &PredicateAPIR{
		ServerFSS{
			db:        db,
			cores:     numCores,
			serverNum: serverNum,
			// one value for the data, four values for the info-theoretic MAC
			fss: fss.ServerInitialize(1 + field.ConcurrentExecutions),
		},
	}
}

func (s *PredicateAPIR) DBInfo() *database.Info {
	return s.ServerFSS.DBInfo()
}

func (s *PredicateAPIR) AnswerBytes(q []byte) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query *query.FSS
	if err := dec.Decode(&query); err != nil {
		return nil, err
	}

	// get answer
	a := s.Answer(query)

	// encode answer
	out := utils.Uint32SliceToByteSlice(a)

	return out, nil
}

func (s *PredicateAPIR) Answer(q *query.FSS) []uint32 {
	out := make([]uint32, 1+field.ConcurrentExecutions)
	tmp := make([]uint32, 1+field.ConcurrentExecutions)

	return s.ServerFSS.answer(q, out, tmp)
}
