package server

import (
	"bytes"
	"encoding/gob"

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

func (a *Amplify) AnswerBytes(qq []byte) ([]byte, error) {
	var mm []*matrix.Matrix
	gob.NewDecoder(bytes.NewBuffer(qq)).Decode(&mm)

	ans := a.Answer(mm)

	// encode
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(ans)

	return buf.Bytes(), nil
}
