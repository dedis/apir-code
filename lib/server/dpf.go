package server

import (
	"bytes"
	"encoding/gob"
	"math/bits"
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

type DPF struct {
	db    *database.DB
	cores int
}

// use variadic argument for cores to achieve backward compatibility
func NewDPF(db *database.DB, cores ...int) *DPF {
	if len(cores) == 0 {
		return &DPF{db: db, cores: runtime.NumCPU()}
	}

	return &DPF{db: db, cores: cores[0]}
}

func (s *DPF) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *DPF) AnswerBytes(q []byte) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query dpf.DPFkey
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

func (s *DPF) Answer(key dpf.DPFkey) []field.Element {
	q := make([]field.Element, s.db.NumColumns*(s.db.BlockSize+1))
	dpf.EvalFullFlatten(key, uint64(bits.Len(uint(s.db.NumColumns)-1)), s.db.BlockSize+1, q)
	return answer(q, s.db)
}
