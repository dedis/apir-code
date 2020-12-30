package server

import (
	db "github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

func NewITVectorGF(db *db.VectorGF) *ITVectorGF {
	return &ITVectorGF{db: db}
}

// ITServer is used to implement the VPIR protocol in the information-theoretic (IT) model
type ITVectorGF struct {
	db *db.VectorGF
}

func (s *ITVectorGF) Answer(q []*field.Element) *field.Element {
	a := field.Zero()
	for i := range s.db.Entries {
		mul := q[i]
		s.db.Entries[i].MulBy(mul)
		a.Add(a, mul)
	}

	return a
}
