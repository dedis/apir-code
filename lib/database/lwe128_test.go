package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/si-co/vpir-code/lib/matrix"
	"github.com/si-co/vpir-code/lib/utils"
)

func TestLWE(t *testing.T) {
	rows, columns := 2048, 2048
	fmt.Println(rows, columns)
	b := make([]byte, rows*columns/8+1)
	rnd := utils.RandomPRG()
	if _, err := rnd.Read(b); err != nil {
		panic("insufficient randomness")
	}

	fmt.Println("randomness generation ok")

	m := matrix.NewBytes(rows, columns)
	for i := 0; i < m.Len(); i++ {
		val := (b[i/8] >> (i % 8)) & 1
		if val >= plaintextModulus {
			panic("Plaintext value too large")
		}
		m.SetData(i, val)
	}

	fmt.Println("db generation ok")

	db := &LWE128{
		Matrix: m,
		Info: Info{
			NumRows:    rows,
			NumColumns: columns,
			BlockSize:  blockSizeLWE,
		},
	}

	fmt.Println("create random matrix A")

	rm := matrix.NewRandom128(
		utils.NewPRG(utils.ParamsDefault128().SeedA),
		utils.ParamsDefault128().N,
		rows)

	fmt.Println("done with create random matrix A")

	fmt.Println("start digest computation BinaryMul")

	ti := time.Now()
	d := matrix.BinaryMul128(rm, db.Matrix)
	fmt.Println("done with digest computation in time:", time.Since(ti).Seconds())

	db.Auth = &Auth{
		Digest: matrix.Matrix128ToBytes(d),
	}

}
