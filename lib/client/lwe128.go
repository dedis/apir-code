package client

import (
	"errors"
	"io"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
	"lukechampine.com/uint128"
)

// LEW128 based authenticated single server PIR client

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
	rand := matrix.NewRandom128(c.rnd, 1, 1)

	// digest is already stored in the state when receiving the database info
	c.state = &StateLWE128{
		A:      matrix.NewRandom128(utils.NewPRG(c.params.SeedA), c.params.N, c.params.L),
		digest: matrix.BytesToMatrix128(c.dbInfo.Digest),
		secret: matrix.NewRandom128(c.rnd, 1, c.params.N),
		i:      i,
		j:      j,
		t:      rand.Get(0, 0),
	}

	// Query has dimension 1 x l
	query := matrix.Mul128(c.state.secret, c.state.A)

	// Error has dimension 1 x l
	e := matrix.NewGauss128(1, c.params.L)

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

func (c *LWE128) Reconstruct(answers *matrix.Matrix128) (uint32, error) {
	s_trans_d := matrix.Mul128(c.state.secret, c.state.digest)
	answers.Sub(s_trans_d)

	outs := make([]uint32, c.params.M)
	for i := 0; i < c.params.M; i++ {
		v := answers.Get(0, i)
		if c.inRange(v) {
			outs[i] = 0
		} else if c.inRange(v.SubWrap(c.state.t)) {
			outs[i] = 1
		} else {
			return 0, errors.New("REJECT")
		}
	}

	return outs[c.state.j], nil
}

func (c *LWE128) ReconstructBytes(a []byte) (uint32, error) {
	return c.Reconstruct(matrix.BytesToMatrix128(a))
}

func (c *LWE128) inRange(val uint128.Uint128) bool {
	// max is q-1, so we add + 1 to B
	tmp := uint128.Max.Sub(uint128.From64(uint64(c.params.B + 1)))
	return val.Cmp64(uint64(c.params.B)) == -1 || val.Cmp(tmp) == 1
}
