package server

/*import (
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic server for scheme working in DB(2^128).
// Both vector and matrix (rebalanced) representations of the
// database are handled by this server, via a boolean variable

// ITSingleGF is the server for the information theoretic single-bit scheme
type ITSingleGF struct {
	rebalanced bool
	db         *database.DB
}

// NewITSingleGF return a server for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITSingleGF(rebalanced bool, db *database.DB) *ITSingleGF {
	return &ITSingleGF{rebalanced: rebalanced, db: db}
}

// Answer computes the answer for the given query
func (s *ITSingleGF) Answer(q []field.Element) []field.Element {
  	var tmp field.Element
	a := make([]field.Element, len(s.db.Entries))
	for i := range s.db.Entries {
		a[i] = field.Zero()
		for j := range s.db.Entries[i] {
			tmp.Mul(&q[j], &s.db.Entries[i][j])
			a[j].Add(&a[j], &tmp)
		}
	}

	return a
}*/
