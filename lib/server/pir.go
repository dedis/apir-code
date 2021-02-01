package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
)

// Information theoretic server for classical PIR scheme working in GF(2).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// PIR is the server for the information theoretic single-bit scheme
type PIR struct {
	db *database.Bytes
}

// NewPIR return a server for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewPIR(db *database.Bytes) *PIR {
	return &PIR{db: db}
}

func (s *PIR) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *PIR) AnswerBytes(q []byte) ([]byte, error) {
	panic("not yet implemented")
	return nil, nil
}

// Answer computes the answer for the given query
func (s *PIR) Answer(q []byte) []byte {
	// Doing simplified scheme if block consists of a single bit
	if s.db.BlockSize == cst.SingleBitBlockLength {
		panic("single-bit classical PIR protocol not implemented")
		//a := make([]byte, s.db.NumRows)
		//for i := 0; i < s.db.NumRows; i++ {
		//for j := 0; j < s.db.NumColumns; j++ {
		//a[i] ^= q[j] & s.db.Entries[i][j]
		//}
		//}
		//return a
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
