package client

import (
	"io"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
)

type Amplify struct {
	t   int  // ECC parameter
	lwe *LWE // base client
}

func NewAmplify(rnd io.Reader, info *database.Info, params *utils.ParamsLWE, t int) *Amplify {
	return &Amplify{
		t:   t,
		lwe: NewLWE(rnd, info, params),
	}
}

func (a *Amplify) Query(i, j int) []*matrix.Matrix {
	query := make([]*matrix.Matrix, a.t+1)

}

func (a *Amplify) Reconstruct(answers []*matrix.Matrix) (uint32, error) {

}
