package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"runtime"

	"github.com/si-co/vpir-code/lib/authfss"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
)

// PIRfss represent the server for the FSS-based complex-queries non-verifiable PIR
type PIRfss struct {
	db    *database.DB
	cores int

	serverNum byte
	fss       *authfss.Fss
}

// NewPIRfss initializes and returns a new server for FSS-based classical PIR
func NewPIRfss(db *database.DB, serverNum byte, cores ...int) *PIRfss {
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &PIRfss{
		db:        db,
		cores:     numCores,
		serverNum: serverNum,
		fss:       authfss.ServerInitialize(1), // only one value for data
	}
}

// DBInfo returns database info
func (s *PIRfss) DBInfo() *database.Info {
	return &s.db.Info
}

// AnswerBytes computes the answer for the given query encoded in bytes
func (s *PIRfss) AnswerBytes(q []byte) ([]byte, error) {
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
func (s *PIRfss) Answer(q *query.FSS) uint32 {
	numIdentifiers := s.db.NumColumns

	out := uint32(0)
	tmp := []uint32{0}

	if !q.And {
		switch q.Target {
		case query.UserId:
			for i := 0; i < numIdentifiers; i++ {
				var id []bool
				email := s.db.KeysInfo[i].UserId.Email
				id, valid := q.IdForEmail(email)
				if !valid {
					continue
				}
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				out = (out + tmp[0]) % field.ModP
			}
			return out
		case query.PubKeyAlgo:
			for i := 0; i < numIdentifiers; i++ {
				id := q.IdForPubKeyAlgo(s.db.KeysInfo[i].PubKeyAlgo)
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				out = (out + tmp[0]) % field.ModP
			}
			return out
		case query.CreationTime:
			for i := 0; i < numIdentifiers; i++ {
				id, err := q.IdForCreationTime(s.db.KeysInfo[i].CreationTime)
				if err != nil {
					panic("impossible to marshal creation date")
				}
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				out = (out + tmp[0]) % field.ModP
			}
			return out
		default:
			panic("not yet implemented")
		}
	} else { // conjunction
		for i := 0; i < numIdentifiers; i++ {
			binaryMatch, err := s.db.KeysInfo[i].CreationTime.MarshalBinary()
			if err != nil {
				panic("impossible to mashal creation date")
			}
			binaryMatch = append(binaryMatch, byte(s.db.KeysInfo[i].PubKeyAlgo))
			id := utils.ByteToBits(binaryMatch)
			s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
			out = (out + tmp[0]) % field.ModP
		}
		return out
	}
}
