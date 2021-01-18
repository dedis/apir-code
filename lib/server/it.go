package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Information theoretic multi-bit server for scheme working in DB(2^128).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this server, via a boolean variable

// ITServer is the server for the information theoretic multi-bit scheme
type ITServer struct {
	db *database.DB
}

// NewITServer return a server for the information theoretic multi-bit scheme,
// working both with the vector and the rebalanced representation of the
// database.
func NewITServer(db *database.DB) *ITServer {
	return &ITServer{db: db}
}

// Answer computes the answer for the given query
func (s *ITServer) Answer(q [][]field.Element) [][]field.Element {
	// Doing simplified scheme if block consists of a single bit
	if s.db.BlockSize == cst.SingleBitBlockLength {
		a := make([][]field.Element, s.db.NumRows)
		for i := 0; i < s.db.NumRows; i++ {
			a[i] = make([]field.Element, 1)
			for j := 0; j < s.db.NumColumns; j++ {
				if s.db.Entries[i][j][0].Equal(&cst.One) {
					a[i][0].Add(&a[i][0], &q[j][0])
				}
			}
		}
		return a
	}

	// parse the query
	qZeroBase := make([]field.Element, s.db.NumColumns)
	qOne := make([][]field.Element, s.db.NumColumns)
	for j := 0; j < s.db.NumColumns; j++ {
		qZeroBase[j] = q[j][0]
		qOne[j] = q[j][1:]
	}

	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	var prodTag field.Element
	m := make([][]field.Element, s.db.NumRows)
	prod := make([]field.Element, s.db.BlockSize)
	// we have to traverse column by column
	for i := 0; i < s.db.NumRows; i++ {
		sumTag := field.Zero()
		sum := field.ZeroVector(s.db.BlockSize)
		m[i] = make([]field.Element, s.db.BlockSize)
		for j := 0; j < s.db.NumColumns; j++ {
			for b := 0; b < s.db.BlockSize; b++ {
				prod[b].Mul(&s.db.Entries[i][j][b], &qZeroBase[j])
				sum[b].Add(&sum[b], &prod[b])

				prodTag.Mul(&s.db.Entries[i][j][b], &qOne[j][b])
				sumTag.Add(&sumTag, &prodTag)
			}
		}
		m[i] = sum
		// add tag
		m[i] = append(m[i], sumTag)
	}

	return m
}
