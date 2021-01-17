package server

import (
	"math/bits"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

func NewDPFServer(db *database.DB) *DPFServer {
	return &DPFServer{db: db}
}

type DPFServer struct {
	db *database.DB
}

func (s *DPFServer) Answer(q dpf.FssKeyEq2P, prfKeys [][]byte, serverNum byte) []field.Element {
	// initialize dpf server
	fServer := dpf.ServerInitialize(prfKeys, uint(bits.Len(uint(constants.DBLength))))

	var tmp field.Element
	a := make([]field.Element, len(s.db.Entries))
	for i := range s.db.Entries {
		a[i] = field.Zero()
		for j := range s.db.Entries[i] {
			tmp.Mul(&s.db.Entries[i][j], fServer.EvaluatePF(serverNum, q, uint(j)))
			a[i].Add(&a[i], &tmp)
		}
	}

	return a
}
