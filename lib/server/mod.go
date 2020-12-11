package server

import (
	"github.com/holiman/uint256"
	db "github.com/si-co/vpir-code/lib/database"
)

// Server represents the server instance in both the IT and C models
type Server interface {
	Answer(q []*uint256.Int) *uint256.Int
}

func NewITServer(db *db.Database) *ITServer {
	return &ITServer{db: db}
}

// ITServer is used to implement the VPIR protocol in the information-theoretic (IT) model
type ITServer struct {
	db *db.Database
}

func (s *ITServer) Answer(q []*uint256.Int) *uint256.Int {
	// Can't use BigZero because it's not deep-copied
	a := uint256.NewInt().SetUint64(0)
	for i := range s.db.Entries {
		mul := uint256.NewInt()
		mul.Mul(s.db.Entries[i], q[i])
		a.Add(a, mul)
	}

	return a
}
