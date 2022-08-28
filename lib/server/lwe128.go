package server

import (
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/matrix"
)

type LWE128 struct {
	db *database.LWE128
}

func NewLWE128(db *database.LWE128) *LWE128 {
	return &LWE128{db: db}
}

func (s *LWE128) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *LWE128) AnswerBytes(q []byte) ([]byte, error) {
	a := s.Answer(matrix.BytesToMatrix128(q))
	return matrix.Matrix128ToBytes(a), nil
}

// Answer function for the LWE-based scheme. The query is represented as a
// vector
func (s *LWE128) Answer(q *matrix.Matrix128) *matrix.Matrix128 {
	return matrix.Mul128(q, s.db.Matrix)
}
