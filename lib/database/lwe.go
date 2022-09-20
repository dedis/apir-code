package database

import (
	"io"

	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
)

type LWE struct {
	Matrix *matrix.MatrixBytes
	Info
}

const plaintextModulus = 2
const blockSizeLWE = 1 // for backward compatibility

func Digest(db *LWE, rows int) *matrix.Matrix {
	return matrix.BinaryMul(
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
	m := matrix.NewBytes(numRows, numColumns)
	// read random bytes for filling out the entries
	data := make([]byte, (numRows*numColumns)/8+1)
	if _, err := rnd.Read(data); err != nil {
		panic(err)
	}

	for i := 0; i < m.Len(); i++ {
		val := (data[i/8] >> (i % 8)) & 1
		if val >= plaintextModulus {
			panic("Plaintext value too large")
		}
		m.SetData(i, val)
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
		DigestLWE: Digest(db, numRows),
	}

	return db
}
