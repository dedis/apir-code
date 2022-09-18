package matrix

import (
	"fmt"
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestEncoding(t *testing.T) {
	m1 := &Matrix{
		rows: 2,
		cols: 2,
		data: []uint32{1, 10, 100, 1000}}

	m2 := &Matrix{
		rows: 2,
		cols: 2,
		data: []uint32{2, 20, 200, 2000}}

	m3 := &Matrix{
		rows: 2,
		cols: 2,
		data: []uint32{3, 30, 300, 3000}}

	mm := []*Matrix{m1, m2, m3}

	b := MatricesToBytes(mm)

	mmRec := BytesToMatrices(b)

	for i := range mmRec {
		require.Equal(t, &mm[i], &mmRec[i])
	}
}

func BenchmarkBinaryMul(b *testing.B) {
	rows, columns := 1024, 1024
	buff := make([]byte, rows*columns/8+1)
	rnd := utils.RandomPRG()
	if _, err := rnd.Read(buff); err != nil {
		panic("insufficient randomness")
	}

	m := NewBytes(rows, columns)
	for i := 0; i < m.Len(); i++ {
		val := (buff[i/8] >> (i % 8)) & 1
		m.SetData(i, val)
	}

	rm := NewRandom(
		utils.NewPRG(utils.ParamsDefault().SeedA),
		utils.ParamsDefault().N,
		rows)

	for i := 0; i < b.N; i++ {
		d := BinaryMul(rm, m)
		// to avoid compiler optimization
		fmt.Println(d.Rows())
	}

}
