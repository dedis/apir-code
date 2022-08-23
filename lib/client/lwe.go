package client

import (
	"errors"
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
	params *utils.ParamsLWE
	rnd    io.Reader
}

type StateLWE struct {
	A      *matrix.Matrix
	digest *matrix.Matrix
	secret *matrix.Matrix
	i      int
	j      int
	t      uint32
}

func NewLWE(rnd io.Reader, info *database.Info, params *utils.ParamsLWE) *LWE {
	return &LWE{
		dbInfo: info,
		params: params,
		rnd:    rnd,
	}
}

func (c *LWE) Query(i, j int) *matrix.Matrix {
	// Lazy way to sample a random scalar
	rand := matrix.NewRandom(c.rnd, 1, 1, MOD)

	// digest is already stored in the state when receiving the database info
	c.state = &StateLWE{
		A:      matrix.NewRandom(utils.NewPRG(c.params.SeedA), c.params.N, c.params.L, MOD),
		digest: matrix.BytesToMatrix(c.dbInfo.Digest),
		secret: matrix.NewRandom(c.rnd, 1, c.params.N, MOD),
		i:      i,
		j:      j,
		t:      rand.Get(0, 0),
	}

	// Query has dimension 1 x l
	query := matrix.Mul(c.state.secret, c.state.A)

	// Error has dimension 1 x l
	e := matrix.NewGauss(c.rnd.(*utils.PRGReader), 1, c.params.L, c.params.Sigma)

	msg := matrix.New(1, c.params.L)
	msg.Set(0, i, c.state.t)

	query.Add(e)
	query.Add(msg)

	return query
}

func (c *LWE) QueryBytes(index int) ([]byte, error) {
	i, j := utils.VectorToMatrixIndices(index, c.dbInfo.NumColumns)
	m := c.Query(i, j)
	return matrix.MatrixToBytes(m), nil
}

func (c *LWE) reconstruct(answers *matrix.Matrix) (uint32, error) {
	s_trans_d := matrix.Mul(c.state.secret, c.state.digest)
	answers.Sub(s_trans_d)

	good := true
	outs := make([]uint32, c.params.M)
	// TODO: shouldn't we break the loop if good == false?
	// or equivalently immediately return a REJECT?
	for i := 0; i < c.params.M; i++ {
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
		return 0, errors.New("REJECT")
	}

	return outs[c.state.j], nil
}

func (c *LWE) ReconstructBytes(a []byte) (uint32, error) {
	return c.reconstruct(matrix.BytesToMatrix(a))
}

func (c *LWE) inRange(val uint32) bool {
	return (val <= c.params.B) || (val >= -c.params.B)
}
