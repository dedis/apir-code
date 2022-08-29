package database

import (
	"io"

	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
	"lukechampine.com/uint128"
)

type LWE128 struct {
	Matrix *matrix.Matrix128
	Info
}

func Digest128(db *LWE128, rows int) *matrix.Matrix128 {
	return matrix.Mul128(
		matrix.NewRandom128(
			utils.NewPRG(utils.ParamsDefault128().SeedA),
			utils.ParamsDefault128().N,
			rows,
		), db.Matrix)
}

func CreateRandomBinaryLWEWithLength128(rnd io.Reader, dbLen int) *LWE128 {
	numRows, numColumns := CalculateNumRowsAndColumns(dbLen, true)
	return CreateRandomBinaryLWE128(rnd, numRows, numColumns)
}

func CreateRandomBinaryLWE128(rnd io.Reader, numRows, numColumns int) *LWE128 {
	m := matrix.New128(numRows, numColumns)
	// read random bytes for filling out the entries
	// For simplicity, we use the whole byte to store 0 or 1
	data := make([]byte, numRows*numColumns)
	if _, err := rnd.Read(data); err != nil {
		panic(err)
	}

	for i := 0; i < numRows; i++ {
		for j := 0; j < numColumns; j++ {
			val := uint128.From64(uint64(data[i] & 1))
			// TODO repristinate this check
			// if val >= plaintextModulus {
			// 	panic("Plaintext value too large")
			// }
			m.Set(i, j, val)
		}
	}

	db := &LWE128{
		Matrix: m,
		Info: Info{
			NumRows:    numRows,
			NumColumns: numColumns,
			BlockSize:  blockSizeLWE,
		},
	}

	db.Auth = &Auth{
		Digest: matrix.Matrix128ToBytes(Digest128(db, numRows)),
	}

	return db
}
