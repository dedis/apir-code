package server

import (
	"math/big"
	"math/bits"

	"github.com/frankw2/libfss/go/libfss"
	"github.com/si-co/vpir-code/lib/constants"
	db "github.com/si-co/vpir-code/lib/database"
)

func NewDPFServer(db *db.Vector) *DPFServer {
	return &DPFServer{db: db}
}

type DPFServer struct {
	db *db.Vector
}

func (s *DPFServer) Answer(q libfss.FssKeyEq2P, prfKeys [][]byte, serverNum byte) *big.Int {
	fServer := libfss.ServerInitialize(prfKeys, uint(bits.Len(uint(constants.DBLength))))
	// Can't use BigZero because it's not deep-copied
	a := big.NewInt(0)
	for i := range s.db.Entries {
		mul := new(big.Int)
		mul.Mul(s.db.Entries[i], big.NewInt(int64(fServer.EvaluatePF(serverNum, q, uint(i)))))
		a.Add(a, mul)
	}

	return a
}
