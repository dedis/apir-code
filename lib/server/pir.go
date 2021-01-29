package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
)

// Information theoretic server for classical PIR scheme working in GF(2).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// ITSingleByte is the server for the information theoretic single-bit scheme
type ITSingleByte struct {
	db *database.Bytes
}

// NewITSingleByte return a server for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITSingleByte(rebalanced bool, db *database.Bytes) *ITSingleByte {
	return &ITSingleByte{db: db}
}

func (s *ITSingleByte) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *ITSingleByte) AnswerBytes(q []byte) ([]byte, error) {
	panic("not yet implemented")
	return nil, nil
}

// Answer computes the answer for the given query
func (s *ITSingleByte) Answer(q []byte) []byte {
	// Doing simplified scheme if block consists of a single bit
	if s.db.BlockSize == cst.SingleBitBlockLength {
		a := make([]byte, s.db.NumRows)
		for i := 0; i < s.db.NumRows; i++ {
			for j := 0; j < s.db.NumColumns; j++ {
				if s.db.Entries[i][j] == byte(1) {
					a[i] ^= q[j]
				}
			}
		}
		return a
	}

	// parse the query
	qZeroBase := make([]byte, s.db.NumColumns)
	for j := 0; j < s.db.NumColumns; j++ {
		qZeroBase[j] = q[j*(s.db.BlockSize+1)]
	}

	m := make([]byte, s.db.NumRows*s.db.BlockSize)
	// we have to traverse column by column
	for i := 0; i < s.db.NumRows; i++ {
		sum := make([]byte, s.db.BlockSize)
		for j := 0; j < s.db.NumColumns; j++ {
			for b := 0; b < s.db.BlockSize; b++ {
				sum[b] ^= s.db.Entries[i][j*s.db.BlockSize+b] & qZeroBase[j]

			}
		}
		copy(m[i*(s.db.BlockSize+1):(i+1)*(s.db.BlockSize+1)-1], sum)
	}

	return m
}
