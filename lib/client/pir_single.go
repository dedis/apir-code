package client

import (
	"crypto/rand"
	"math"

	cst "github.com/si-co/vpir-code/lib/constants"
	"golang.org/x/crypto/blake2b"
)

// Information theoretic classical PIR client for scheme working in GF(2).
// Both vector and matrix (rebalanced) representations of the database are
// handled by this client, via a boolean variable

// Client for the information theoretic classical PIR single-bit scheme
type ITSingleByte struct {
	xof        blake2b.XOF
	state      *itSingleByteState
	rebalanced bool
}

type itSingleByteState struct {
	ix       int
	iy       int // unused if not rebalanced
	dbLength int
}

// NewItSingleByte return a client for the classical PIR single-bit scheme in
// Byte(2), working both with the vector and the rebalanced representation of the
// database.
func NewITSingleByte(xof blake2b.XOF, rebalanced bool) *ITSingleByte {
	return &ITSingleByte{
		xof:        xof,
		rebalanced: rebalanced,
		state:      nil,
	}
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITSingleByte) Query(index int, numServers int) [][]byte {
	//if invalidQueryInputs(index, numServers) {
	//	panic("invalid query inputs")
	//}

	// set the client state depending on the db representation
	switch c.rebalanced {
	case false:
		// iy is unused if the database is represented as a vector
		c.state = &itSingleByteState{
			ix:       index,
			dbLength: cst.DBLength,
		}
	case true:
		// verified at server side if integer square
		dbLengthSqrt := int(math.Sqrt(cst.DBLength))
		ix := index % dbLengthSqrt
		iy := index / dbLengthSqrt

		c.state = &itSingleByteState{
			ix:       ix,
			iy:       iy,
			dbLength: dbLengthSqrt,
		}
	}

	vectors, err := c.secretSharing(numServers)
	if err != nil {
		panic(err)
	}

	return vectors
}

func (c *ITSingleByte) Reconstruct(answers [][]byte) (byte, error) {
	answersLen := len(answers[0])
	sum := make([]byte, answersLen)

	// sum answers
	for i := 0; i < answersLen; i++ {
		for s := range answers {
			sum[i] ^= answers[s][i]
		}
	}

	// select index depending on the matrix representation
	i := 0
	if c.rebalanced {
		i = c.state.iy
	}

	return sum[i], nil
}

func (c *ITSingleByte) secretSharing(numServers int) ([][]byte, error) {
	ei := make([]byte, c.state.dbLength)
	ei[c.state.ix] = byte(1)

	vectors := make([][]byte, numServers)

	// create query vectors for all the servers
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]byte, c.state.dbLength)
	}

	zero := byte(0)

	// for all except one server, we need dbLength random elements
	// to perform the secret sharing
	b := make([]byte, c.state.dbLength*(numServers-1))
	_, err := rand.Read(b)
	if err != nil {
		panic("error in randomness generation")
	}
	for i := 0; i < c.state.dbLength; i++ {
		sum := zero
		for k := 0; k < numServers-1; k++ {
			rand := b[c.state.dbLength*k+i] % 2
			vectors[k][i] = rand
			sum ^= rand
		}
		vectors[numServers-1][i] = ei[i] ^ sum
	}

	return vectors, nil
}
