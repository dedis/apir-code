package server

import "github.com/si-co/vpir-code/lib/database"

type LWE struct {
	db *database.LWE
}

func NewLWE(db *database.LWE) *LWE {
	return &LWE{db: db}
}

func (s *LWE) DBInfo() *database.Info {
	return &s.db.Info
}

// TODO LWE implement AnswerBytes
func (s *LWE) AnswerBytes(q []byte) ([]byte, error) {
	panic("not yet implemented")
}

// Answer function for the LWE-based scheme. The query is represented as a
// vector and takes therefore the same type as the database
func (s *LWE) Answer(q *database.LWE) *database.LWE {
	return database.Mul(q, s.db)
}
