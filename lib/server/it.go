package server

import (
	"encoding/binary"
	"runtime"

	"github.com/si-co/vpir-code/lib/database"
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
	n := len(q) / (8 * 2)
	data := make([]uint32, n)

	for i := 0; i < n; i++ {
		memIndex := i * 8 * 2

		data[i] = uint32{
			binary.LittleEndian.Uint64(q[memIndex : memIndex+8]),
			binary.LittleEndian.Uint64(q[memIndex+8 : memIndex+16]),
		}
	}

	// get answer
	a := s.Answer(data)

	res := make([]byte, len(a)*8*2)
	for k := 0; k < len(a); k++ {
		binary.LittleEndian.PutUint64(res[k*8*2:k*8*2+8], a[k][0])
		binary.LittleEndian.PutUint64(res[k*8*2+8:k*8*2+8+8], a[k][1])
	}

	return res, nil
}

// Answer computes the answer for the given query
func (s *IT) Answer(q []uint32) []uint32 {
	return answer(q, s.db, s.cores)
}
