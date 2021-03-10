package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
)

// Information theoretic server for classical PIR scheme working in GF(2).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// PIR is the server for the information theoretic classical PIR scheme
type PIR struct {
	db *database.Bytes
}

// NewPIR return a server for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewPIR(db *database.Bytes) *PIR {
	if db.BlockSize == cst.SingleBitBlockLength {
		panic("single-bit classical PIR protocol not implemented")
	}
	return &PIR{db: db}
}

// DBInfo returns database info
func (s *PIR) DBInfo() *database.Info {
	return &s.db.Info
}

// AnswerBytes computes the answer for the given query encoded in bytes
func (s *PIR) AnswerBytes(q []byte) ([]byte, error) {
	return s.Answer(q), nil
}

// Answer computes the answer for the given query
func (s *PIR) Answer(q []byte) []byte {
	return answerPIR(q, s.db)
}
