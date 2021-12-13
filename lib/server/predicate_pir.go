package server

import (
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

// PredicatePIR represent the server for the FSS-based complex-queries unauthenticated PIR
type PredicatePIR struct {
	*serverFSS
}

// NewPredicatePIR initializes and returns a new server for FSS-based classical PIR
func NewPredicatePIR(db *database.DB, serverNum byte, cores ...int) *PredicatePIR {
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &PredicatePIR{
		&serverFSS{
			db:        db,
			cores:     numCores,
			serverNum: serverNum,
			fss:       fss.ServerInitialize(1), // only one value for data
		},
	}
}

// DBInfo returns database info
func (s *PredicatePIR) DBInfo() *database.Info {
	return s.serverFSS.dbInfo()
}

// AnswerBytes computes the answer for the given query encoded in bytes
func (s *PredicatePIR) AnswerBytes(q []byte) ([]byte, error) {
	out := make([]uint32, 1)
	tmp := make([]uint32, 1)

	return s.serverFSS.answerBytes(q, out, tmp)
}

// Answer computes the answer for the given query
func (s *PredicatePIR) Answer(q *query.FSS) uint32 {
	out := []uint32{0}
	tmp := []uint32{0}

	return s.serverFSS.answer(q, out, tmp)[0]
}
