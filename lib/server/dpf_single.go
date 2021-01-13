package server

import (
	"math/bits"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

func NewDPFServer(db *database.GF) *DPFServer {
	return &DPFServer{db: db}
}

type DPFServer struct {
	db *database.GF
}

func (s *DPFServer) Answer(q []dpf.FssKeyEq2P, prfKeys [][]byte, serverNum byte, blockSize int) []field.Element {
	// initialize dpf server
	fServer := dpf.ServerInitialize(prfKeys, uint(bits.Len(uint(constants.DBLength))))

	// Doing simplified scheme if block consists of a single bit
	if blockSize == cst.SingleBitBlockLength {
		a := make([]field.Element, 1)
		a[0].SetZero()
		for i := range s.db.Entries {
			for j := range s.db.Entries[i] {
				if s.db.Entries[i][j].Equal(&cst.One) {
					a[i].Add(&a[i], fServer.EvaluatePF(serverNum, q[0], uint(j)))
				}
			}
		}

		return a
	}

	return nil

	// Part for the multi-bit scheme
}
