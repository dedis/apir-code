package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic multi-bit server for scheme working in GF(2^128).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// ITMulti is the server for the information theoretic multi-bit scheme
type ITMulti struct {
	rebalanced bool
	db         *database.GF
}

// NewITMulti return a server for the information theoretic multi-bit scheme,
// working both with the vector and the rebalanced representation of the
// database.
func NewITMulti(rebalanced bool, db *database.GF) *ITMulti {
	return &ITMulti{rebalanced: rebalanced, db: db}
}

// Answer computes the answer for the given query
func (s *ITMulti) Answer(q [][]field.Element, blockSize int) []field.Element {
	// Doing simplified scheme if block consists of one element
	if blockSize == cst.BlockSizeSingleBit {
		//var tmp field.Element
		a := make([]field.Element, len(s.db.Entries))
		for i := range s.db.Entries {
			a[i] = field.Zero()
			for j := range s.db.Entries[i] {
				//tmp.Mul(&q[j][0], &s.db.Entries[i][j])
				//a[i].Add(&a[i], &tmp)
				if s.db.Entries[i][j].Equal(&cst.One) {
					a[i].Add(&a[i], &q[j][0])
				}
			}
		}
		return a
	}

	// parse the query
	qZeroBase := make([]field.Element, cst.DBLength)
	qOne := make([][]field.Element, cst.DBLength)
	for i := range q {
		qZeroBase[i] = q[i][0]
		qOne[i] = q[i][1:]
	}

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
			prod.Mul(&s.db.Entries[j][i], &qZeroBase[j])
			sum.Add(&sum, &prod)

			prodTag.Mul(&s.db.Entries[j][i], &qOne[j][i])
			sumTag.Add(&sumTag, &prodTag)
		}
		m[i] = sum
		tag.Add(&tag, &sumTag)
	}

	// add tag
	m = append(m, tag)

	return m
}
