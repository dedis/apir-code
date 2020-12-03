package server

import (
	db "github.com/si-co/vpir-code/lib/database"
	"math/big"
)

func CreateServer(db *db.Database) server {
	return server{db: db}
}

type server struct {
	db *db.Database
}

func (s server) Answer(q []*big.Int) *big.Int {
	// Can't use BigZero because it's not deep-copied
	a := big.NewInt(0)
	for i := range s.db.Entries {
		mul := new(big.Int)
		mul.Mul(s.db.Entries[i], q[i])
		a.Add(a, mul)
	}

	return a
}
