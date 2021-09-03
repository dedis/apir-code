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
// TODO: refactor this function
func (s *FSS) Answer(q *query.FSS) []uint32 {
	numIdentifiers := s.db.NumColumns
	switch q.Target {
	case query.UserId:
		out := make([]uint32, s.db.BlockSize)
		tmp := make([]uint32, s.db.BlockSize)
		for i := 0; i < numIdentifiers; i++ {
			var id []bool
			if q.FromStart != 0 {
				id = utils.ByteToBits([]byte(s.db.KeysInfo[i].UserId.Email[:q.FromStart]))
			} else if q.FromEnd != 0 {
				id = utils.ByteToBits([]byte(s.db.KeysInfo[i].UserId.Email[len(s.db.KeysInfo[i].UserId.Email)-q.FromEnd:]))
			} else {
				id = utils.ByteToBits([]byte(s.db.KeysInfo[i].UserId.Email))
			}
			s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
			for i := range out {
				out[i] = (out[i] + tmp[i]) % field.ModP
			}
		}
		return out
	case query.PubKeyAlgo:
		out := make([]uint32, s.db.BlockSize)
		tmp := make([]uint32, s.db.BlockSize)
		for i := 0; i < numIdentifiers; i++ {
			id := utils.ByteToBits([]byte{uint8(s.db.KeysInfo[i].PubKeyAlgo)})
			s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
			for i := range out {
				out[i] = (out[i] + tmp[i]) % field.ModP
			}
		}
		return out
	default:
		panic("not yet implemented")
	}

}
