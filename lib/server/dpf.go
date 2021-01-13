package server

import (
	"math/bits"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

func NewDPF(db *database.GF) *DPF {
	return &DPF{db: db}
}

type DPF struct {
	db *database.GF
}

func (s *DPF) Answer(q []dpf.FssKeyEq2P, prfKeys [][]byte, serverNum byte, blockSize int) []field.Element {
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

	// Part for the multi-bit scheme

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in GF(2^128)^b are executed component-wise
	m := make([]field.Element, blockSize)
	tag := field.Zero()

	var prod, prodTag field.Element
	// we have to traverse column by column
	for i := 0; i < blockSize; i++ {
		sum := field.Zero()
		sumTag := field.Zero()
		for j := 0; j < cst.DBLength; j++ {
			prod.Mul(&s.db.Entries[j][i], fServer.EvaluatePF(serverNum, q[0], uint(j)))
			sum.Add(&sum, &prod)

			prodTag.Mul(&s.db.Entries[j][i], fServer.EvaluatePF(serverNum, q[i+1], uint(j)))
			sumTag.Add(&sumTag, &prodTag)
		}
		m[i] = sum
		tag.Add(&tag, &sumTag)
	}

	// add tag
	m = append(m, tag)

	return m

}
