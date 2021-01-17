package server

import (
	"github.com/si-co/vpir-code/lib/database"
)

// Information theoretic server for classical PIR scheme working in DB(2).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// ITSingleByte is the server for the information theoretic single-bit scheme
type ITSingleByte struct {
	rebalanced bool
	db         *database.Bytes
}

// NewITSingleByte return a server for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITSingleByte(rebalanced bool, db *database.Bytes) *ITSingleByte {
	return &ITSingleByte{rebalanced: rebalanced, db: db}
}

// Answer computes the answer for the given query
func (s *ITSingleByte) Answer(q []byte) []byte {
	a := make([]byte, len(s.db.Entries))
	for i := range s.db.Entries {
		for j := range s.db.Entries[i] {
			mul := s.db.Entries[i][j] & q[j]
			a[i] ^= mul
		}

	}

	return a
}
