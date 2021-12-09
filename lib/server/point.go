package server

import (
	"runtime"

	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/database"
)

// PIR is the server for the information theoretic classical PIR scheme
// working in GF(2).
// This scheme is used for point queries, both in the authenticated and
// unauthenticated setting. The former adds Merkle-tree based authentication
// information to the database entries, but this is trasparent from the server
// perspective and only changes the database creation.
type PIR struct {
	db    *database.Bytes
	cores int
}

// NewPIR return a server for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewPIR(db *database.Bytes, cores ...int) *PIR {
	if len(cores) == 0 {
		return &PIR{db: db, cores: runtime.NumCPU()}
	}
	return &PIR{db: db, cores: cores[0]}
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
	nRows := s.db.NumRows
	nCols := s.db.NumColumns

	var prevPos, nextPos int
	out := make([]byte, nRows*s.db.BlockSize)

	for i := 0; i < nRows; i++ {
		for j := 0; j < nCols; j++ {
			nextPos += s.db.BlockLengths[i*nCols+j]
		}
		xorValues(
			s.db.Entries[prevPos:nextPos],
			s.db.BlockLengths[i*nCols:(i+1)*nCols],
			q,
			s.db.BlockSize,
			out[i*s.db.BlockSize:(i+1)*s.db.BlockSize])
		prevPos = nextPos
	}
	return out
}

// XORs entries and q block by block of size bl
func xorValues(entries []byte, blockLens []int, q []byte, bl int, out []byte) {
	pos := 0
	for j := range blockLens {
		if (q[j/8]>>(j%8))&1 == byte(1) {
			fastxor.Bytes(out, out, entries[pos:pos+blockLens[j]])
		}
		pos += blockLens[j]
	}
}
