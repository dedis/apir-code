package matrix

import (
	"testing"

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
