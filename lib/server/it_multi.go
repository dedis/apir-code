package server

import (
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

}
