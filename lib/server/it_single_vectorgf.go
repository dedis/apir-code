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
	// Can't use BigZero because it's not deep-copied
	a := field.NewUint64(0)
	for i := range s.db.Entries {
		mul := field.NewUint64(1)
		mul.Mul(s.db.Entries[i], q[i])
		a.Add(a, mul)
	}

	return a
}
