package server

import (
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
)

// PredicateAPIR represent the server for the FSS-based complex-queries authenticated PIR
type PredicateAPIR struct {
	*serverFSS
}

func NewPredicateAPIR(db *database.DB, serverNum byte, cores ...int) *PredicateAPIR {
	// use variadic argument for cores to achieve backward compatibility
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &PredicateAPIR{
		&serverFSS{
			db:        db,
			cores:     numCores,
			serverNum: serverNum,
			// one value for the data, four values for the info-theoretic MAC
			fss: fss.ServerInitialize(1 + field.ConcurrentExecutions),
		},
	}
}

func (s *PredicateAPIR) DBInfo() *database.Info {
	return s.serverFSS.dbInfo()
}

func (s *PredicateAPIR) AnswerBytes(q []byte) ([]byte, error) {
	out := make([]uint32, 1+field.ConcurrentExecutions)
	tmp := make([]uint32, 1+field.ConcurrentExecutions)

	return s.serverFSS.answerBytes(q, out, tmp)
}

func (s *PredicateAPIR) Answer(q *query.FSS) []uint32 {
	out := make([]uint32, 1+field.ConcurrentExecutions)
	tmp := make([]uint32, 1+field.ConcurrentExecutions)

	return s.serverFSS.answer(q, out, tmp)
}
