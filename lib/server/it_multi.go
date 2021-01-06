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

	// TODO: need to change the matrix logic for the tag

	// parse the query
	qZeroBase := make([]field.Element, constants.DBLength)
	qOne := make([][]field.Element, constants.DBLength)

	for i := range q {
		qZeroBase[i] = q[i][0]
		qOne[i] = q[i][1:]
	}

	// extend qZeroBase
	qZero := make([][]field.Element, blockLength)
	for i := range qZero {
		qZero[i] = qZeroBase
	}

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in GF(2^128)^b are executed component-wise
	//m := make([]lib.Element, blockLength)
	//t := make([]lib.Element, blockLength)

	return nil
}
