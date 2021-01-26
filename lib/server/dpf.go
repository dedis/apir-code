package server

import (
	"bytes"
	"encoding/gob"
	"math/bits"
	"sync"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

type DPF struct {
	db        *database.DB
	serverNum byte
	mu        sync.Mutex
}

func NewDPF(db *database.DB, serverNum byte) *DPF {
	return &DPF{db: db,
		serverNum: serverNum,
	}
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
	s.mu.Lock()
	defer s.mu.Unlock()
	q := make([]field.Element, s.db.NumColumns*(s.db.BlockSize+1))
	dpf.EvalFullFlatten(key, uint64(bits.Len(uint(s.db.NumColumns))), s.db.BlockSize+1, q)

	return answer(q, s.db)
}
