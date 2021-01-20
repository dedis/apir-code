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
func answer(q [][]field.Element, db *database.DB) [][]field.Element {
	// Doing simplified scheme if block consists of a single bit
	if db.BlockSize == cst.SingleBitBlockLength {
		a := make([][]field.Element, db.NumRows)
		for i := 0; i < db.NumRows; i++ {
			a[i] = make([]field.Element, 1)
			for j := 0; j < db.NumColumns; j++ {
				if db.Entries[i][j][0].Equal(&cst.One) {
					a[i][0].Add(&a[i][0], &q[j][0])
				}
			}
		}
		return a
	}

	// parse the query
	qZeroBase := make([]field.Element, db.NumColumns)
	qOne := make([][]field.Element, db.NumColumns)
	for j := 0; j < db.NumColumns; j++ {
		qZeroBase[j] = q[j][0]
		qOne[j] = q[j][1:]
	}

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	var prodTag field.Element
	m := make([][]field.Element, db.NumRows)
	prod := make([]field.Element, db.BlockSize)
	// we have to traverse column by column
	for i := 0; i < db.NumRows; i++ {
		sumTag := field.Zero()
		sum := field.ZeroVector(db.BlockSize)
		m[i] = make([]field.Element, db.BlockSize)
		for j := 0; j < db.NumColumns; j++ {
			for b := 0; b < db.BlockSize; b++ {
				prod[b].Mul(&db.Entries[i][j][b], &qZeroBase[j])
				sum[b].Add(&sum[b], &prod[b])

				prodTag.Mul(&db.Entries[i][j][b], &qOne[j][b])
				sumTag.Add(&sumTag, &prodTag)
			}
		}
		m[i] = sum
		// add tag
		m[i] = append(m[i], sumTag)
	}

	return m
}