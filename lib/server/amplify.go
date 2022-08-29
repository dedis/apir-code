package server

import (
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/matrix"
)

type Amplify struct {
	lwe *LWE
}

func NewAmplify(db *database.LWE) *Amplify {
	return &Amplify{
		lwe: NewLWE(db),
	}
}

func (a *Amplify) DBInfo() *database.Info {
	return &a.lwe.db.Info
}

// TODO: run in parallel?
func (a *Amplify) Answer(qq []*matrix.Matrix) []*matrix.Matrix {
	ans := make([]*matrix.Matrix, len(qq))
	for i, q := range qq {
		ans[i] = matrix.Mul(q, a.lwe.db.Matrix)
	}

	return ans
}
