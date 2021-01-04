package client

import (
	"errors"
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"golang.org/x/crypto/blake2b"
)

// Information theoretic client for scheme working in GF(2^128).
// Both vector and matrix (rebalanced) representations of the
// database are handled by this client, via a boolean variable

// Client for the information theoretic single-bit scheme
type ITSingleGF struct {
	xof        blake2b.XOF
	state      *itSingleGFState
	rebalanced bool
	vpir       bool
}

type itSingleGFState struct {
	ix       int
	iy       int // unused if not rebalanced
	alpha    *field.Element
	dbLength int
}

// NewItSingleGF return a client for the information theoretic single-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewITSingleGF(xof blake2b.XOF, rebalanced, vpir bool) *ITSingleGF {
	return &ITSingleGF{
		xof:        xof,
		rebalanced: rebalanced,
		vpir:       vpir,
		state:      nil,
	}
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *ITSingleGF) Query(index int, numServers int) [][]*field.Element {
	if invalidQueryInputs(index, numServers) {
		panic("invalid query inputs")
	}

	// sample random alpha using blake2b
	alpha := field.One()
	if vpir {
		alpha := field.RandomXOF(c.xof)
	}

	// set the client state depending on the db representation
	switch c.rebalanced {
	case false:
		// iy is unused if the database is represented as a vector
		c.state = &itSingleGFState{
			ix:       index,
			alpha:    alpha,
			dbLength: cst.DBLength,
		}
	case true:
		// verified at server side if integer square
		dbLengthSqrt := int(math.Sqrt(cst.DBLength))
		ix := index % dbLengthSqrt
		iy := index / dbLengthSqrt

		c.state = &itSingleGFState{
			ix:       ix,
			iy:       iy,
			alpha:    alpha,
			dbLength: dbLengthSqrt,
		}
	}

	vectors, err := c.secretSharing(numServers)
	if err != nil {
		panic(err)
	}

	return vectors
}

func (c *ITSingleGF) Reconstruct(answers [][]*field.Element) (*field.Element, error) {
	answersLen := len(answers[0])
	sum := make([]*field.Element, answersLen)

	// sum answers
	for i := 0; i < answersLen; i++ {
		sum[i] = field.Zero()
		for s := range answers {
			sum[i].Add(sum[i], answers[s][i])
		}

		if !sum[i].Equal(c.state.alpha) && !sum[i].Equal(field.Zero()) {
			return nil, errors.New("REJECT!")
		}
	}

	// select index depending on the matrix representation
	i := 0
	if c.rebalanced {
		i = c.state.iy
	}

	switch {
	case sum[i].Equal(c.state.alpha):
		return constants.One, nil
	case sum[i].Equal(constants.Zero):
		return constants.Zero, nil
	default:
		return nil, errors.New("REJECT!")
	}
}

func (c *ITSingleGF) secretSharing(numServers int) ([][]*field.Element, error) {
	eialpha := make([]*field.Element, c.state.dbLength)
	vectors := make([][]*field.Element, numServers)

	// create query vectors for all the servers
	for k := 0; k < numServers; k++ {
		vectors[k] = make([]*field.Element, c.state.dbLength)
	}

	zero := field.Zero()
	// for all except one server, we need dbLength random elements
	// to perform the secret sharing
	numRandomElements := c.state.dbLength * (numServers - 1)
	randomElements := field.RandomVectorXOF(numRandomElements, c.xof)
	for i := 0; i < c.state.dbLength; i++ {
		// create basic vector
		eialpha[i] = zero

		// set alpha at the index we want to retrieve
		if i == c.state.ix {
			eialpha[i] = c.state.alpha
		}

		// create k - 1 random vectors
		sum := field.Zero()
		for k := 0; k < numServers-1; k++ {
			rand := randomElements[c.state.dbLength*k+i]
			vectors[k][i] = rand
			sum.Add(sum, rand)
		}
		vectors[numServers-1][i] = field.Zero()
		vectors[numServers-1][i].Add(eialpha[i], sum)
	}

	return vectors, nil
}

// return true if the query inputs are invalid
func invalidQueryInputs(index int, numServers int) bool {
	return (index < 0 || index > cst.DBLength) && numServers < 1
}
