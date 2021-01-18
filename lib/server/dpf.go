package server

import (
	"log"
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
	db        *database.DB
	serverNum byte
}

func (s *DPF) Answer(q dpf.DPFkey, serverNum byte) [][]field.Element {
	// Doing simplified scheme if block consists of a single bit
	if s.db.BlockSize == cst.SingleBitBlockLength {
		log.Fatal("Not implemented")
		return nil
		/*
			a := make([]field.Element, 1)
			a[0].SetZero()
			for j := range s.db.Entries {
				for j := range s.db.Entries[j] {
					if s.db.Entries[j][j].Equal(&cst.One) {
						eval := fServer.EvaluatePFVector(serverNum, q, uint(j))
						a[j].Add(&a[j], eval[0])
					}
				}
			}

			return a
		*/
	}

	// Part for the multi-bit scheme

	// evaluate dpf
	dpfOut := make([][]field.Element, s.db.NumColumns)
	for j := 0; j < len(dpfOut); j++ {
		dpfOut[j] = make([]field.Element, s.db.BlockSize+1)
	}

	dpf.EvalFull(q, uint64(bits.Len(uint(s.db.NumColumns))), dpfOut)

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
			for b := 0; b < s.db.BlockSize; b++ {
				prod[b].Mul(&s.db.Entries[i][j][b], &dpfOut[j][0])
				sum[b].Add(&sum[b], &prod[b])

				prodTag.Mul(&s.db.Entries[i][j][b], &dpfOut[j][b+1])
				sumTag.Add(&sumTag, &prodTag)
			}
		}
		m[i] = sum
		// add tag
		m[i] = append(m[i], sumTag)
	}

	return m

}
