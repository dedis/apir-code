package server

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
)

type serverFSS struct {
	db    *database.DB
	cores int

	serverNum byte
	fss       *fss.Fss
}

func (s *serverFSS) dbInfo() *database.Info {
	return &s.db.Info
}

func (s *serverFSS) answerBytes(q []byte, out, tmp []uint32) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query *query.FSS
	if err := dec.Decode(&query); err != nil {
		return nil, err
	}

	// get answer
	a := s.answer(query, out, tmp)

	return utils.Uint32SliceToByteSlice(a), nil
}

func (s *serverFSS) answer(q *query.FSS, out, tmp []uint32) []uint32 {
	numIdentifiers := s.db.NumColumns

	if !q.And && !q.Avg && !q.Sum {
		switch q.Target {
		case query.UserId:
			for i := 0; i < numIdentifiers; i++ {
				email := s.db.KeysInfo[i].UserId.Email
				id, valid := q.IdForEmail(email)
				if !valid {
					continue
				}
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				for i := range out {
					out[i] = (out[i] + tmp[i]) % field.ModP
				}
			}
			return out
		case query.PubKeyAlgo:
			for i := 0; i < numIdentifiers; i++ {
				id := q.IdForPubKeyAlgo(s.db.KeysInfo[i].PubKeyAlgo)
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				for i := range out {
					out[i] = (out[i] + tmp[i]) % field.ModP
				}
			}
			return out
		case query.CreationTime:
			for i := 0; i < numIdentifiers; i++ {
				id, err := q.IdForCreationTime(s.db.KeysInfo[i].CreationTime)
				if err != nil {
					panic("impossible to marshal creation date")
				}
				s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
				for i := range out {
					out[i] = (out[i] + tmp[i]) % field.ModP
				}
			}
			return out
		default:
			panic("not yet implemented")
		}
	} else if q.And && !q.Avg && !q.Sum { // conjunction
		for i := 0; i < numIdentifiers; i++ {
			// year
			yearMatch, err := q.IdForYearCreationTime(s.db.KeysInfo[i].CreationTime)
			if err != nil {
				panic(err)
			}
			// edu
			email := s.db.KeysInfo[i].UserId.Email
			id, valid := q.IdForEmail(email)
			if !valid {
				continue
			}
			in := append(yearMatch, id...)
			s.fss.EvaluatePF(s.serverNum, q.FssKey, in, tmp)
			for i := range out {
				out[i] = (out[i] + tmp[i]) % field.ModP
			}
		}
		return out

	} else if q.And && q.Sum && !q.Avg { // sum
		for i := 0; i < numIdentifiers; i++ {
			binaryMatch, err := s.db.KeysInfo[i].CreationTime.MarshalBinary()
			if err != nil {
				panic("impossible to marshal creation date")
			}
			binaryMatch = append(binaryMatch, byte(s.db.KeysInfo[i].PubKeyAlgo))
			id := utils.ByteToBits(binaryMatch)

			s.fss.EvaluatePF(s.serverNum, q.FssKey, id, tmp)
			for i := range out {
				out[i] = (out[i] + tmp[i]*uint32(s.db.KeysInfo[i].BitLength)) % field.ModP
			}
		}

		return out
	} else if q.And && q.Avg && !q.Sum { // avg
		now := time.Now()
		for i := 0; i < numIdentifiers; i++ {
			// year
			yearMatch, err := q.IdForYearCreationTime(s.db.KeysInfo[i].CreationTime)
			if err != nil {
				panic(err)
			}
			in := yearMatch
			s.fss.EvaluatePF(s.serverNum, q.FssKey, in, tmp)
			for i := range out {
				// COUNT
				out[i] = (out[i] + tmp[i]) % field.ModP

				// SUM
				diffYears := uint32(now.Sub(s.db.KeysInfo[i].CreationTime).Seconds() / 31207680) // TODO: round
				out[i] = (out[i] + (tmp[i]*diffYears)%field.ModP) % field.ModP
			}
		}
		// TODO: need to implement modular inverse to divide sum by count
		return out
	} else {
		panic("query not recognized")
	}
}
