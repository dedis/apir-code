package server

import (
	"bytes"
	"encoding/gob"
	"math/bits"
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/fss"
)

type FSS struct {
	db    *database.DB
	cores int
}

// use variadic argument for cores to achieve backward compatibility
func NewFSS(db *database.DB, cores ...int) *FSS {
	if len(cores) == 0 {
		return &FSS{db: db, cores: runtime.NumCPU()}
	}

	return &FSS{db: db, cores: cores[0]}
}

func (s *FSS) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *FSS) AnswerBytes(q []byte) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query FSS.FSSkey
	if err := dec.Decode(&query); err != nil {
		return nil, err
	}

	// get answer
	a := s.Answer(query)

	// encode answer
	buf.Reset()
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(a); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *FSS) Answer(key fss.FssKeyEq2P) []uint32 {
	q := make([]uint32, s.db.NumColumns*(s.db.BlockSize+1))
	FSS.EvalFullFlatten(key, uint64(bits.Len(uint(s.db.NumColumns)-1)), s.db.BlockSize+1, q)
	return answer(q, s.db, s.cores)
}
