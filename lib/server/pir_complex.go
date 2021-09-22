package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

// PIRfss represent the server for the FSS-based complex-queries non-verifiable PIR
type PIRfss struct {
	db    *database.DB
	cores int

	serverNum byte
	fss       *fss.Fss
}

// NewPIRfss initializes and returns a new server for FSS-based classical PIR
func NewPIRfss(db *database.DB, serverNum byte, prfKeys [][]byte, cores ...int) *PIRfss {
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &PIRfss{
		db:        db,
		cores:     numCores,
		serverNum: serverNum,
		// one value for the data, four values for the info-theoretic MAC
		fss: fss.ServerInitialize(prfKeys),
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
	binary.BigEndian.PutUint32(out, uint32(a))

	return out, nil
}

// Answer computes the answer for the given query
func (s *PIRfss) Answer(q *query.FSS) int {
	numIdentifiers := s.db.NumColumns

	out := 0

	if !q.And {
		switch q.Target {
		case query.UserId:
			for i := 0; i < numIdentifiers; i++ {
				var id []bool
				email := s.db.KeysInfo[i].UserId.Email
				if q.FromStart != 0 {
					if q.FromStart > len(email) {
						continue
					}
					id = utils.ByteToBits([]byte(email[:q.FromStart]))
				} else if q.FromEnd != 0 {
					if q.FromEnd > len(email) {
						continue
					}
					id = utils.ByteToBits([]byte(email[len(email)-q.FromEnd:]))
				} else {
					h := blake2b.Sum256([]byte(email))
					id = utils.ByteToBits(h[:16])
				}
				fmt.Println(s.serverNum, out)
				out += s.fss.EvaluatePF(s.serverNum, q.FssKey, id)
			}
			return out
		case query.PubKeyAlgo:
			for i := 0; i < numIdentifiers; i++ {
				id := utils.ByteToBits([]byte{uint8(s.db.KeysInfo[i].PubKeyAlgo)})
				out += s.fss.EvaluatePF(s.serverNum, q.FssKey, id)
			}
			return out
		case query.CreationTime:
			for i := 0; i < numIdentifiers; i++ {
				binaryMatch, err := s.db.KeysInfo[i].CreationTime.MarshalBinary()
				if err != nil {
					panic("impossible to marshal creation date")
				}
				id := utils.ByteToBits(binaryMatch)
				out += s.fss.EvaluatePF(s.serverNum, q.FssKey, id)
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
			out += s.fss.EvaluatePF(s.serverNum, q.FssKey, id)

		}
		return out
	}
}
