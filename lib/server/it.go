package server

import (
	"bytes"
	"encoding/binary"
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic multi-bit server for scheme working in DB(2^128).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// IT is the server for the information theoretic multi-bit scheme
type IT struct {
	db    *database.DB
	cores int
}

// NewIT return a server for the information theoretic multi-bit scheme,
// working both with the vector and the rebalanced representation of the
// database.
func NewIT(db *database.DB, cores ...int) *IT {
	if len(cores) == 0 {
		return &IT{db: db, cores: runtime.NumCPU()}
	}
	return &IT{db: db, cores: cores[0]}
}

func (s *IT) DBInfo() *database.Info {
	return &s.db.Info
}

// AnswerBytes decode the input, execute Answer and encodes the output
func (s *IT) AnswerBytes(q []byte) ([]byte, error) {
	res := make([]uint32, len(q)/field.Bytes)
	buf := bytes.NewReader(q)
	err := binary.Read(buf, binary.BigEndian, &res)
	if err != nil {
		return nil, err
	}

	// get answer
	a := s.Answer(res)

	out := make([]byte, len(a)*field.Bytes)
	buff := new(bytes.Buffer)
	err = binary.Write(buff, binary.BigEndian, a)
	if err != nil {
		return nil, err
	}
	out = buff.Bytes()

	return out, nil
}

// Answer computes the answer for the given query
func (s *IT) Answer(q []uint32) []uint32 {
	return answer(q, s.db, s.cores)
}
