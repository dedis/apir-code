package server

import (
	"math/big"

	db "github.com/si-co/vpir-code/lib/database"
)

// ITServer is used to implement the VPIR protocol in the information-theoretic (IT) model
type ITMatrixServer struct {
	db *db.Matrix
}

func NewITMatrixServer(db *db.Matrix) *ITMatrixServer {
	return &ITMatrixServer{db: db}
}

func (s *ITMatrixServer) Answer(q []*big.Int) []*big.Int {
	a := make([]*big.Int, len(s.db.Entries[0]))
	// Can't use BigZero because it's not deep-copied
	for i := range s.db.Entries {
		a[i] = big.NewInt(0)
		for j := range s.db.Entries[i] {
			mul := new(big.Int)
			mul.Mul(s.db.Entries[i][j], q[j])
			a[i].Add(a[i], mul)
		}

	}

	return a
}
