package server

import (
	"bytes"
	"encoding/gob"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic multi-bit server for scheme working in DB(2^128).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// IT is the server for the information theoretic multi-bit scheme
type IT struct {
	db *database.DB
}

// NewIT return a server for the information theoretic multi-bit scheme,
// working both with the vector and the rebalanced representation of the
// database.
func NewIT(db *database.DB) *IT {
	return &IT{db: db}
}

func (s *IT) DBInfo() *database.Info {
	return &s.db.Info
}

// AnswerBytes decode the input, execute Answer and encodes the output
func (s *IT) AnswerBytes(q []byte) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query []field.Element
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

// Answer computes the answer for the given query
func (s *IT) Answer(q []field.Element) []field.Element {
	return answer(q, s.db)
}
