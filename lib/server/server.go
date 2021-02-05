package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Server is a scheme-agnostic VPIR server interface, implemented by both IT
// and DPF-based schemes
type Server interface {
	AnswerBytes([]byte) ([]byte, error)
	DBInfo() *database.Info
}

// Answer computes the answer for the given query
func answer(q []field.Element, db *database.DB) []field.Element {
	// Doing simplified scheme if block consists of a single bit
	if db.BlockSize == cst.SingleBitBlockLength {
		a := make([]field.Element, db.NumRows)
		for i := 0; i < db.NumRows; i++ {
			for j := 0; j < db.NumColumns; j++ {
				if db.Entries[i][j].Equal(&cst.One) {
					a[i].Add(&a[i], &q[j])
				}
			}
		}
		return a
	}

	// parse the query
	qZeroBase := make([]field.Element, db.NumColumns)
	qOne := make([]field.Element, db.NumColumns*db.BlockSize)
	for j := 0; j < db.NumColumns; j++ {
		qZeroBase[j] = q[j*(db.BlockSize+1)]
		copy(qOne[j*db.BlockSize:(j+1)*db.BlockSize], q[j*(db.BlockSize+1)+1:(j+1)*(db.BlockSize+1)])
	}

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	var prodTag, prod field.Element
	m := make([]field.Element, db.NumRows*(db.BlockSize+1))
	// we have to traverse column by column
	for i := 0; i < db.NumRows; i++ {
		sumTag := field.Zero()
		sum := field.ZeroVector(db.BlockSize)
		for j := 0; j < db.NumColumns; j++ {
			for b := 0; b < db.BlockSize; b++ {
				if db.Entries[i][j*db.BlockSize+b].IsZero() {
					// no need to multiply is the the element value is zero
					continue
				}
				prod.Mul(&db.Entries[i][j*db.BlockSize+b], &qZeroBase[j])
				sum[b].Add(&sum[b], &prod)

				prodTag.Mul(&db.Entries[i][j*db.BlockSize+b], &qOne[j*db.BlockSize+b])
				sumTag.Add(&sumTag, &prodTag)
			}
		}
		copy(m[i*(db.BlockSize+1):(i+1)*(db.BlockSize+1)-1], sum)
		// add tag
		m[(i+1)*(db.BlockSize+1)-1].Set(&sumTag)
	}

	return m
}
