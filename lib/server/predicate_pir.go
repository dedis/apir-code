package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

// PredicatePIR represent the server for the FSS-based complex-queries unauthenticated PIR
type PredicatePIR struct {
	ServerFSS
}

// NewPredicatePIR initializes and returns a new server for FSS-based classical PIR
func NewPredicatePIR(db *database.DB, serverNum byte, cores ...int) *PredicatePIR {
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &PredicatePIR{
		ServerFSS{
			db:        db,
			cores:     numCores,
			serverNum: serverNum,
			fss:       fss.ServerInitialize(1), // only one value for data
		},
	}
}

// DBInfo returns database info
func (s *PredicatePIR) DBInfo() *database.Info {
	return s.ServerFSS.DBInfo()
}

// AnswerBytes computes the answer for the given query encoded in bytes
func (s *PredicatePIR) AnswerBytes(q []byte) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query *query.FSS
	if err := dec.Decode(&query); err != nil {
		return nil, err
	}

	a := s.Answer(query)

	// encode answer
	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, a)

	return out, nil
}

// Answer computes the answer for the given query
func (s *PredicatePIR) Answer(q *query.FSS) uint32 {
	out := []uint32{0}
	tmp := []uint32{0}

	return s.ServerFSS.answer(q, out, tmp)[0]
}
