package client

import (
	"github.com/si-co/vpir-code/lib/field"
	"golang.org/x/crypto/blake2b"
)

// Information theoretic client for multi-bit scheme working in
// GF(2^128). Both vector and matrix (rebalanced) representations of
// the database are handled by this client, via a boolean variable

// ITMulti represents the client for the information theoretic multi-bit scheme
type ITMulti struct {
	xof        blake2b.XOF
	state      *itMultiState
	rebalanced bool
}

type itMultiState struct {
	ix       int
	iy       int // unused if not rebalanced
	alpha    field.Element
	dbLength int
}

// NewITSingleGF return a client for the information theoretic multi-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITSingleGF(xof blake2b.XOF, rebalanced bool) *ITSingleGF {
	return &ITSingleGF{
		xof:        xof,
		rebalanced: rebalanced,
		state:      nil,
	}
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITMulti) Query(index int, numServers int) [][]field.Element {
	// TODO: check query inputs

	// sample random alpha using blake2b
	alpha := field.RandomXOF(c.xof)

	// compute vector a = (alpha, alpha^2, ..., alpha^b)

	// compute vector a = (alpha, alpha^2, ..., alpha^b)
}
