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
	"golang.org/x/crypto/blake2b"
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
		fss:       fss.ServerInitialize(prfKeys, 64, 2*field.ConcurrentExecutions),
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

// TODO: refactor this function
func (s *FSS) Answer(q *query.FSS) []uint32 {
	numIdentifiers := s.db.NumColumns

	out := make([]uint32, s.db.BlockSize*field.ConcurrentExecutions)
	tmp := make([]uint32, s.db.BlockSize*field.ConcurrentExecutions)

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
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				for i := range out {
					out[i] = (out[i] + tmp[i]) % field.ModP
				}
			}
			return out
		case query.PubKeyAlgo:
			for i := 0; i < numIdentifiers; i++ {
				id := utils.ByteToBits([]byte{uint8(s.db.KeysInfo[i].PubKeyAlgo)})
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				for i := range out {
					out[i] = (out[i] + tmp[i]) % field.ModP
				}
			}
			return out
		case query.CreationTime:
			for i := 0; i < numIdentifiers; i++ {
				binaryMatch, err := s.db.KeysInfo[i].CreationTime.MarshalBinary()
				if err != nil {
					panic("impossible to marshal creation date")
				}
				id := utils.ByteToBits(binaryMatch)
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				for i := range out {
					out[i] = (out[i] + tmp[i]) % field.ModP
				}
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
			for i := range out {
				out[i] = (out[i] + tmp[i]) % field.ModP
			}
		}
		return out

	}

}
