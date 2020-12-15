package server

import (
	"math/big"

	db "github.com/si-co/vpir-code/lib/database"
)

// Server represents the server instance in both the IT and C models
type Server interface {
	Answer(q []*big.Int) *big.Int
}

func NewITServer(db *db.Vector) *ITServer {
	return &ITServer{db: db}
}

// ITServer is used to implement the VPIR protocol in the information-theoretic (IT) model
type ITServer struct {
	db *db.Vector
}

func (s *ITServer) Answer(q []*big.Int) *big.Int {
	// Can't use BigZero because it's not deep-copied
	a := big.NewInt(0)
	for i := range s.db.Entries {
		mul := new(big.Int)
		mul.Mul(s.db.Entries[i], q[i])
		a.Add(a, mul)
	}

	return a
}
