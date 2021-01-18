package server

import (
	"math/bits"

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

func (s *DPF) Answer(q dpf.DPFkey, serverNum byte, blockSize int) [][]field.Element {
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

	// evaluate dpf
	dpfOut := make([][]field.Element, s.db.NumRows)
	for i := 0; i < len(dpfOut); i++ {
		dpfOut[i] = make([]field.Element, s.db.BlockSize+1)
	}

	dpf.EvalFull(q, uint64(bits.Len(uint(s.db.NumRows))), dpfOut)

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in GF(2^128)^b are executed component-wise
	var prodTag field.Element
	m := make([][]field.Element, s.db.NumRows)
	prod := make([]field.Element, s.db.BlockSize)
	for i := 0; i < s.db.NumRows; i++ {
		sumTag := field.Zero()
		sum := field.ZeroVector(s.db.BlockSize)
		m[i] = make([]field.Element, s.db.BlockSize)
		for j := 0; j < s.db.NumColumns; j++ {
			for b := 0; b < blockSize; b++ {
				prod[b].Mul(&s.db.Entries[i][j][b], &dpfOut[i][0])
				sum[b].Add(&sum[b], &prod[b])

				prodTag.Mul(&s.db.Entries[i][j][b], &dpfOut[i][b+1])
				sumTag.Add(&sumTag, &prodTag)
			}
		}
		m[i] = sum
		// add tag
		m[i] = append(m[i], sumTag)
	}

	return m

}
