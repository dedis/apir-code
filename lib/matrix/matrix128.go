package matrix

import (
	"encoding/binary"
	"io"

	"github.com/si-co/vpir-code/lib/utils"
	"lukechampine.com/uint128"
)

type Matrix128 struct {
	rows int
	cols int
	data []uint128.Uint128
}

func New128(r int, c int) *Matrix128 {
	return &Matrix128{
		rows: r,
		cols: c,
		data: make([]uint128.Uint128, r*c),
	}
}

func Matrix128ToBytes(in *Matrix128) []byte {
	// we first store rows and cols to allow reconstruction
	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, uint32(in.rows))
	c := make([]byte, 4)
	binary.BigEndian.PutUint32(c, uint32(in.cols))
	params := append(r, c...)
	out := make([]byte, len(in.data)*16)
	for i := range in.data {
		in.data[i].PutBytes(out[i*16 : (i+1)*16])
	}
	// finally we store the data and append the params in front of them
	return append(params, out...)
}

func BytesToMatrix128(in []byte) *Matrix128 {
	// retrieve the matrix dimensions
	rows := int(binary.BigEndian.Uint32(in[:4]))
	cols := int(binary.BigEndian.Uint32(in[4:]))
	data := make([]uint128.Uint128, len(in[8:])/16)
	for i := range data {
		data[i] = uint128.FromBytes(in[8+i*16:])
	}
	return &Matrix128{
		rows: rows,
		cols: cols,
		data: data,
	}
}

func NewRandom128(rnd io.Reader, r int, c int) *Matrix128 {
	bytesMod := utils.ParamsDefault128().BytesMod
	b := make([]byte, bytesMod*r*c)
	if _, err := rnd.Read(b); err != nil {
		panic(err)
	}

	m := New128(r, c)

	for i := 0; i < len(m.data); i++ {
		m.data[i] = uint128.FromBytes(b[i*bytesMod : (i+1)*bytesMod])
	}

	return m
}

func NewGauss128(r int, c int) *Matrix128 {
	m := New128(r, c)
	for i := 0; i < len(m.data); i++ {
		g := utils.GaussSample()
		if g >= 0 {
			m.data[i] = uint128.From64(uint64(g))
		} else {
			m.data[i] = uint128.Max.Sub64(uint64(-g))
		}
	}

	return m
}

func (m *Matrix128) Set(r int, c int, v uint128.Uint128) {
	m.data[m.cols*r+c] = v
}

func (m *Matrix128) Get(r int, c int) uint128.Uint128 {
	return m.data[m.cols*r+c]
}

func (m *Matrix128) Rows() int {
	return m.rows
}

func (m *Matrix128) Cols() int {
	return m.cols
}

func BinaryMul128(a *Matrix128, b *MatrixBytes) *Matrix128 {
	if a.cols != b.rows {
		panic("Dimension mismatch")
	}

	out := New128(a.rows, b.cols)
	for i := 0; i < a.rows; i++ {
		for k := 0; k < a.cols; k++ {
			for j := 0; j < b.cols; j++ {
				if b.data[b.cols*k+j] != byte(0) {
					out.data[b.cols*i+j] = out.data[b.cols*i+j].AddWrap(a.data[a.cols*i+k])
				}
				// if (b.Data[(b.cols*k+j)/8]>>(b.cols*k+j)%8)&1 != byte(0) {
				// 	out.data[b.cols*i+j] = out.data[b.cols*i+j].AddWrap(a.data[a.cols*i+k])
				// }
			}
		}
	}

	return out
}

func Mul128(a *Matrix128, b *Matrix128) *Matrix128 {
	if a.cols != b.rows {
		panic("Dimension mismatch")
	}

	out := New128(a.rows, b.cols)
	for i := 0; i < a.rows; i++ {
		for k := 0; k < a.cols; k++ {
			for j := 0; j < b.cols; j++ {
				tmp := a.data[a.cols*i+k].MulWrap(b.data[b.cols*k+j])
				out.data[b.cols*i+j] = out.data[b.cols*i+j].AddWrap(tmp)
			}
		}
	}

	return out
}

func (a *Matrix128) Add(b *Matrix128) {
	if a.cols != b.cols || a.rows != b.rows {
		panic("Dimension mismatch")
	}

	for i := 0; i < len(a.data); i++ {
		a.data[i] = a.data[i].AddWrap(b.data[i])
	}
}

func (a *Matrix128) Sub(b *Matrix128) {
	if a.cols != b.cols || a.rows != b.rows {
		panic("Dimension mismatch")
	}

	for i := 0; i < len(a.data); i++ {
		a.data[i] = a.data[i].SubWrap(b.data[i])
	}
}
