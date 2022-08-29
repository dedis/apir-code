package database

import (
	"github.com/si-co/vpir-code/lib/ecc"
	"github.com/si-co/vpir-code/lib/matrix"
)

type Amplify struct {
	t      int // ECC parameter
	Matrix *matrix.Matrix
	Info
}

const plaintextModulusAmplify = 2
const blockSizeAmplify = 1 // for backward compatibility

// TODO: need to have a function with rows and cols and one with length?
func DigestAmplify(t int, data []uint32) []*matrix.Matrix {
	// initialize ECC
	ecc := ecc.New(t)

	// encode database
	dataAmplified := make([]uint32, len(data)*(t+1))
	// encode all the database
	for i := range data {
		copy(dataAmplified[i*(t+1):(i+1)*(t+1)], ecc.Encode(data[i]))
	}

	// compute digests
	digests := make([]*matrix.Matrix, t+1)
	zj := make([]uint32, 0, len(data))
	for j := 0; j < t+1; j++ {
		for i := range data {
			zj = append(zj, dataAmplified[i*(t+1)+j])
		}

		digests[j] = Digest(&LWE{
			Matrix: matrix.NewWithData(len(zj), 1, zj),
			Info:   Info{}}, len(zj),
		)

		zj = make([]uint32, 0, len(data))
	}

	return digests
}
