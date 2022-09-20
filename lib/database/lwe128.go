package database

import (
	"io"

	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
)

type LWE128 struct {
	Matrix *matrix.MatrixBytes
	Info
}

func Digest128(db *LWE128, rows int) *matrix.Matrix128 {
	return matrix.BinaryMul128(
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
	m := matrix.NewBytes(numRows, numColumns)
	// read random bytes for filling out the entries
	// the +1 takes into account a float division by 8
	data := make([]byte, numRows*numColumns/8+1)
	if _, err := rnd.Read(data); err != nil {
		panic(err)
	}

	for i := 0; i < m.Len(); i++ {
		val := (data[i/8] >> (i % 8)) & 1
		if val >= plaintextModulus {
			panic("value too large")
		}
		m.SetData(i, val)
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
		DigestLWE128: Digest128(db, numRows),
	}

	return db
}
