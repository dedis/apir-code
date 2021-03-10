package server

import (
	"bytes"
	"encoding/gob"
	"math/bits"

	"github.com/dimakogan/dpf-go/dpf"
	"github.com/si-co/vpir-code/lib/database"
)

// DPF-based server for classical PIR scheme working in GF(2).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// PIRdpf is the server for the PIR-based classical PIR scheme
type PIRdpf struct {
	db        *database.Bytes
}

// NewPIRdpf initializes and returns a new server for DPF-based classical PIR
func NewPIRdpf(db *database.Bytes) *PIRdpf {
	return &PIRdpf{db: db}
}

// DBInfo returns database info
func (s *PIRdpf) DBInfo() *database.Info {
	return &s.db.Info
}

// AnswerBytes computes the answer for the given query encoded in bytes
func (s *PIRdpf) AnswerBytes(q []byte) ([]byte, error) {
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

// Answer computes the answer for the given query
func (s *PIRdpf) Answer(key dpf.DPFkey) []byte {
	q := dpf.EvalFull(key, uint64(bits.Len(uint(s.db.NumColumns)-1)))
	return answerPIR(q, s.db)
}
