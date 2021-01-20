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

// ITServer is the server for the information theoretic multi-bit scheme
type ITServer struct {
	db *database.DB
}

// NewITServer return a server for the information theoretic multi-bit scheme,
// working both with the vector and the rebalanced representation of the
// database.
func NewITServer(db *database.DB) *ITServer {
	return &ITServer{db: db}
}

func (s *ITServer) DBInfo() *database.Info {
	return &s.db.Info
}

// AnswerBytes decode the input, execute Answer and encodes the output
func (s *ITServer) AnswerBytes(q []byte) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query [][]field.Element
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
func (s *ITServer) Answer(q [][]field.Element) [][]field.Element {
	return answer(q, s.db)
}
