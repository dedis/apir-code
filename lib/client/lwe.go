package client

import (
	"io"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
)

// LEW based authenticated single server PIR client

// Ciphertext modulus
const MOD = 1 << 32

// Client description
type LWE struct {
	dbInfo *database.Info
	state  *StateLWE
	params *ParamsLWE
	rnd    io.Reader
}

type StateLWE struct {
	digest *matrix.Matrix
	secret *matrix.Matrix
	i      int
	j      int
	t      uint32
}

type ParamsLWE struct {
	p     uint32  // plaintext modulus
	n     int     // lattice/secret dimension
	sigma float64 // Error parameter

	l int    // number of rows of database
	m int    // number of columns of database
	B uint32 // bound used in reconstruction

	A *matrix.Matrix // matrix  used to generate digest
}

func ParamsDefault() *ParamsLWE {
	p := &ParamsLWE{
		p:     2,
		n:     1024,
		sigma: 6.0,
		l:     512,
		m:     128,
		B:     1000,
	}

	p.A = matrix.NewRandom(utils.NewPRG(utils.GetDefaultSeedMatrixA()), p.n, p.l, MOD)
	return p
}

func NewLWE(info *database.Info) *LWE {
	return &LWE{
		dbInfo: info,
		params: ParamsDefault(),
	}
}

func (c *LWE) query(i, j int) *matrix.Matrix {
	// Lazy way to sample a random scalar
	rand := matrix.NewRandom(c.rnd, 1, 1, MOD)

	// digest is already stored in the state when receiving the database info
	state := &StateLWE{
		secret: matrix.NewRandom(c.rnd, 1, c.params.n, MOD),
		i:      i,
		j:      j,
		t:      rand.Get(0, 0),
	}

	// Query has dimension 1 x l
	query := matrix.Mul(state.secret, c.params.A)

	// Error has dimension 1 x l
	e := matrix.NewGauss(1, c.params.l, c.params.sigma)

	msg := matrix.New(1, c.params.l)
	msg.Set(0, i, state.t)

	query.Add(e)
	query.Add(msg)

	return query
}

func (c *LWE) QueryBytes(index int) ([]byte, error) {
	panic("not yet implemented")
}

func (c *LWE) reconstruct(answers *matrix.Matrix) uint32 {
	s_trans_d := matrix.Mul(c.state.secret, c.state.digest)
	answers.Sub(s_trans_d)

	good := true
	outs := make([]uint32, c.params.m)
	for i := 0; i < c.params.m; i++ {
		v := answers.Get(0, i)
		if c.inRange(v) {
			outs[i] = 0
		} else if c.inRange(v - c.state.t) {
			outs[i] = 1
		} else {
			good = false
		}
	}

	if !good {
		panic("Incorrect reconstruction")
	}

	return outs[c.state.j]
}

func (c *LWE) ReconstructBytes(a []byte) ([]uint64, error) {
	panic("not yet implemented")
}

func (c *LWE) inRange(val uint32) bool {
	return (val <= c.params.B) || (val >= -c.params.B)
}
