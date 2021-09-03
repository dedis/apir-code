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

type FSS struct {
	db    *database.DB
	cores int

	serverNum byte
	fss       *fss.Fss
}

// use variadic argument for cores to achieve backward compatibility
func NewFSS(db *database.DB, serverNum byte, prfKeys [][]byte, cores ...int) *FSS {
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &FSS{
		db:        db,
		cores:     numCores,
		serverNum: serverNum,
		fss:       fss.ServerInitialize(prfKeys, 64, db.BlockSize),
	}

}

func (s *FSS) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *FSS) AnswerBytes(q []byte) ([]byte, error) {
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

// TODO: how to do here? It is quite strange that the server imports the client
// Define the query to be outside of the function?
func (s *FSS) Answer(q *query.FSS) []uint32 {
	numIdentifiers := s.db.NumColumns
	//qEval := make([]uint32, (s.db.BlockSize+1)*numIdentifiers)
	out := make([]uint32, s.db.BlockSize)
	switch q.QueryType {
	case query.KeyId:
		//tmp := make([]uint32, s.db.BlockSize+1)
		tmp := make([]uint32, 2)
		for i := 0; i < numIdentifiers; i++ {
			id := utils.ByteToBits([]byte(s.db.KeysInfo[i].UserId.Email))
			s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
			out[0] = (out[0] + tmp[0]) % field.ModP
			out[1] = (out[1] + tmp[1]) % field.ModP

			//copy(qEval[i*(s.db.BlockSize+1):(i+1)*(s.db.BlockSize+1)], tmp)
		}
		//return answer(qEval, s.db, s.cores)
		return out
	default:
		panic("not yet implemented")
	}

}
