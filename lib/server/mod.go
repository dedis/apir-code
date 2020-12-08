package server

import (
	"github.com/ncw/gmp"
	db "github.com/si-co/vpir-code/lib/database"
)

// Server represents the server instance in both the IT and C models
type Server interface {
	Answer(q []*gmp.Int) *gmp.Int
}

func NewITServer(db *db.Database) *ITServer {
	return &ITServer{db: db}
}

// ITServer is used to implement the VPIR protocol in the information-theoretic (IT) model
type ITServer struct {
	db *db.Database
}

func (s *ITServer) Answer(q []*gmp.Int) *gmp.Int {
	// Can't use BigZero because it's not deep-copied
	a := gmp.NewInt(0)
	mul := gmp.NewInt(0)
	for i := range s.db.Entries {
		mul.Mul(s.db.Entries[i], q[i])
		//fmt.Printf("%d ", mul)
		//fmt.Printf("Adding %d and %d ", a, mul)
		a.Add(a, mul)
		//fmt.Printf("Result: %d\n", a)
	}
	//fmt.Printf("Server result: %d\n", a.Int64())

	return a
}
