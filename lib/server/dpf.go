package server

import (
	"math/bits"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

func NewDPF(db *database.DB) *DPF {
	return &DPF{db: db}
}

type DPF struct {
	db *database.DB
}

func (s *DPF) Answer(q dpf.DPFkey, serverNum byte, blockSize int) []field.Element {
	// Doing simplified scheme if block consists of a single bit
	if blockSize == cst.SingleBitBlockLength {
    panic("Not implemented")
    /*
		a := make([]field.Element, 1)
		a[0].SetZero()
		for i := range s.db.Entries {
			for j := range s.db.Entries[i] {
				if s.db.Entries[i][j].Equal(&cst.One) {
					eval := fServer.EvaluatePFVector(serverNum, q, uint(j))
					a[i].Add(&a[i], eval[0])
				}
			}
		}

		return a
    */
	}

	// Part for the multi-bit scheme

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in GF(2^128)^b are executed component-wise
	m := make([]field.Element, blockSize)
	tag := field.Zero()

  dpfOut := make([][]field.Element, cst.DBLength)
  for i := 0; i < len(dpfOut); i++ {
    dpfOut[i] = make([]field.Element, blockSize+1)
  }

	dpf.EvalFull(q, uint64(bits.Len(uint(constants.DBLength))), dpfOut)

  var tmp field.Element
  for i := 0; i < len(dpfOut); i++ {
    for j := 0; j < blockSize; j++ {
      tmp.Mul(&s.db.Entries[i][j], &dpfOut[i][0])
      m[j].Add(&m[j], &tmp)

      tmp.Mul(&s.db.Entries[i][j], &dpfOut[i][j+1])
      tag.Add(&tag, &tmp)
    }
  }

	// add tag
	m = append(m, tag)

	return m

}
