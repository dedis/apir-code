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

func Digest(db *LWE, rows int) *matrix.Matrix {
	return matrix.Mul(
		matrix.NewRandom(
			utils.NewPRG(utils.ParamsDefault().SeedA),
			utils.ParamsDefault().N,
			rows,
		), db.Matrix)
}

func CreateRandomBinaryLWEWithLength(rnd io.Reader, dbLen int) *LWE {
	numRows, numColumns := CalculateNumRowsAndColumns(dbLen, true)
	return CreateRandomBinaryLWE(rnd, numRows, numColumns)
}

func CreateRandomBinaryLWE(rnd io.Reader, numRows, numColumns int) *LWE {
	m := matrix.New(numRows, numColumns)
	// read random bytes for filling out the entries
	// For simplicity, we use the whole byte to store 0 or 1
	data := make([]byte, numRows*numColumns)
	if _, err := rnd.Read(data); err != nil {
		panic(err)
	}

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
		Digest: matrix.MatrixToBytes(Digest(db, numRows)),
	}

	return db
}
