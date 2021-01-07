package server

import (
	"github.com/si-co/vpir-code/lib/constants"
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
func (s *ITMulti) Answer(q [][]field.Element) []field.Element {
	blockLength := constants.BlockLength

	// parse the query
	qZeroBase := make([]field.Element, constants.DBLength)
	qOne := make([][]field.Element, constants.DBLength)
	for i := range q {
		qZeroBase[i] = q[i][0]
		qOne[i] = q[i][1:]
	}

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in GF(2^128)^b are executed component-wise
	m := make([]field.Element, blockLength)
	tag := field.Zero()

	// we have to traverse column by column
	for i := 0; i < blockLength; i++ {
		sum := field.Zero()
		sumTag := field.Zero()
		for j := 0; j < constants.DBLength; j++ {
			prod := field.Mul(s.db.Entries[j][i], qZeroBase[j])
			sum.AddTo(&prod)

			prodTag := field.Mul(s.db.Entries[j][i], qOne[j][i])
			sumTag.AddTo(&prodTag)
		}
		m[i] = sum
		tag.AddTo(&sumTag)
	}

	// add tag
	m = append(m, tag)

	return m
}
