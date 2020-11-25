package main

import (
	"math/big"
)

type Server struct {
}

func (s Server) Answer(db *Database, q []*big.Int) *big.Int {
	// TODO: why bigZero doesn't work?
	a := big.NewInt(0)
	for i := range db.Entries {
		mul := new(big.Int)
		mul.Mul(db.Entries[i], q[i])
		a.Add(a, mul)
	}

	return a
}
