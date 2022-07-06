package client

import (
	"io"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
	"lukechampine.com/uint128"
)

// LEW based authenticated single server PIR client

// Client description
type LWE128 struct {
	dbInfo *database.Info
	state  *StateLWE128
	params *utils.ParamsLWE
	rnd    io.Reader
}

type StateLWE128 struct {
	A      *matrix.Matrix128
	digest *matrix.Matrix128
	secret *matrix.Matrix128
	i      int
	j      int
	t      uint128.Uint128
}

func NewLWE128(rnd io.Reader, info *database.Info, params *utils.ParamsLWE) *LWE128 {
	return &LWE128{
		dbInfo: info,
		params: params,
		rnd:    rnd,
	}
}

func (c *LWE128) Query(i, j int) *matrix.Matrix128 {
	// Lazy way to sample a random scalar
	rand := matrix.NewRandom128(c.rnd, 1, 1, uint128.Max)

	// digest is already stored in the state when receiving the database info
	c.state = &StateLWE128{
		A:      matrix.NewRandom128(utils.NewPRG(c.params.SeedA), c.params.N, c.params.L, uint128.Max),
		digest: matrix.BytesToMatrix128(c.dbInfo.Digest),
		secret: matrix.NewRandom128(c.rnd, 1, c.params.N, uint128.Max),
		i:      i,
		j:      j,
		t:      rand.Get(0, 0),
	}

	// Query has dimension 1 x l
	query := matrix.Mul128(c.state.secret, c.state.A)

	// Error has dimension 1 x l
	e := matrix.NewGauss128(1, c.params.L, c.params.Sigma)

	msg := matrix.New128(1, c.params.L)
	msg.Set(0, i, c.state.t)

	query.Add(e)
	query.Add(msg)

	return query
}

func (c *LWE128) QueryBytes(index int) ([]byte, error) {
	i, j := utils.VectorToMatrixIndices(index, c.dbInfo.NumColumns)
	m := c.Query(i, j)
	return matrix.Matrix128ToBytes(m), nil
}

func (c *LWE128) reconstruct(answers *matrix.Matrix128) uint32 {
	s_trans_d := matrix.Mul128(c.state.secret, c.state.digest)
	answers.Sub(s_trans_d)

	good := true
	outs := make([]uint32, c.params.M)
	for i := 0; i < c.params.M; i++ {
		v := answers.Get(0, i)
		if c.inRange(v) {
			outs[i] = 0
		} else if c.inRange(v.Sub(c.state.t)) {
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

func (c *LWE128) ReconstructBytes(a []byte) (uint32, error) {
	return c.reconstruct(matrix.BytesToMatrix128(a)), nil
}

// TODO: check how to set B
func (c *LWE128) inRange(val uint128.Uint128) bool {
	if val.Cmp(uint128.From64(uint64(c.params.B))) != 1 {
		return false
	} else if val.Cmp(uint128.Max.Sub(uint128.From64(uint64(c.params.B)))) != -1 {
		return false
	} else {
		return true
	}
}
