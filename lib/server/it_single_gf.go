package server

import (
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic server for scheme working in GF(2^128).
// Both vector and matrix (rebalanced) representations of the
// database are handled by this server, via a boolean variable

// ITSingleGF is the server for the information theoretic single-bit scheme
type ITSingleGF struct {
	rebalanced bool
	// one of the two databases is unused, depending on
	// rebalanced
	dbVector *database.VectorGF
	dbMatrix *database.MatrixGF
}

// NewITSingleGF return a server for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITSingleGF(rebalanced bool, db interface{}) *ITSingleGF {
	s := &ITSingleGF{rebalanced: rebalanced}
	switch rebalanced {
	case false:
		s.dbVector = db.(*database.VectorGF)
	case true:
		s.dbMatrix = db.(*database.MatrixGF)
	}

	return s
}

// Answer computes the answer for the given query
func (s *ITSingleGF) Answer(q []*field.Element) []*field.Element {
	// TODO: db always as a matrix with only one row if vector?
	switch s.rebalanced {
	case false:
		a := []*field.Element{field.Zero()}
		for i := range s.dbVector.Entries {
			mul := q[i]
			s.dbVector.Entries[i].MulBy(mul)
			a[0].Add(a[0], mul)
		}

		return a
	case true:
		a := make([]*field.Element, len(s.dbMatrix.Entries[0]))
		for i := range s.dbMatrix.Entries {
			a[i] = field.Zero()
			for j := range s.dbMatrix.Entries[i] {
				mul := q[j]
				s.dbMatrix.Entries[i][j].MulBy(q[j])
				a[i].Add(a[i], mul)
			}

		}

		return a
	default:
		return nil
	}

}
