package database

import (
	"io"

	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
)

type LWE struct {
	Matrix *matrix.Matrix
	Info
}

const plaintextModulus = 2
const blockSizeLWE = 1 // for backward compatibility

func DefaultDigest(db *LWE) *matrix.Matrix {
	return matrix.Mul(
		matrix.NewRandom(
			utils.NewPRG(utils.ParamsDefault().SeedA),
			utils.ParamsDefault().N,
			utils.ParamsDefault().L,
			utils.ParamsDefault().Mod,
		), db.Matrix)
}

func CreateZeroLWE(numRows, numColumns int) *LWE {
	m := matrix.New(numRows, numColumns)

	db := &LWE{
		Matrix: m,
		Info: Info{
			NumRows:    numRows,
			NumColumns: numColumns,
			BlockSize:  blockSizeLWE,
		},
	}

	db.Auth.Digest = matrix.MatrixToBytes(DefaultDigest(db))

	return db
}

func CreateRandomBinaryLWE(rnd io.Reader, numRows, numColumns int) *LWE {
	m := matrix.New(numRows, numColumns)
	// read random bytes for filling out the entries
	// For simplicity, we use the whole byte to store 0 or 1
	data := make([]byte, numRows*numColumns)
	if _, err := rnd.Read(data); err != nil {
		panic(err)
	}

	// TODO: create function to generate bits db
	for i := 0; i < numRows; i++ {
		for j := 0; j < numColumns; j++ {
			val := uint32(data[i] & 1)
			if val >= plaintextModulus {
				panic("Plaintext value too large")
			}
			m.Set(i, j, val)
		}
	}

	db := &LWE{
		Matrix: m,
		Info: Info{
			NumRows:    numRows,
			NumColumns: numColumns,
			BlockSize:  blockSizeLWE,
		},
	}

	db.Auth = &Auth{
		Digest: matrix.MatrixToBytes(DefaultDigest(db)),
	}

	return db
}
