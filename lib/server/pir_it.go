package server

import (
	"github.com/lukechampine/fastxor"
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
	bs := s.db.BlockSize
	m := make([]byte, s.db.NumRows*bs)
	// we have to traverse column by column
	for i := 0; i < s.db.NumRows; i++ {
		sum := make([]byte, bs)
		for j := 0; j < s.db.NumColumns; j++ {
			if q[j] == byte(1) {
				fastxor.Bytes(sum, sum, s.db.Entries[i][j*bs:bs*(j+1)])
			}
		}
		copy(m[i*bs:(i+1)*bs], sum)
	}

	return m
}
