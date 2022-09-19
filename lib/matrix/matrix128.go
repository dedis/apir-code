package matrix

import (
	"encoding/binary"
	"io"

	"github.com/si-co/vpir-code/lib/utils"
	"lukechampine.com/uint128"
)

/*
#cgo CFLAGS: -std=c99 -O3 -march=native -msse4.1 -maes -mavx2 -mavx
#include <matrix.h>
*/
import "C"

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

	aa := make([]byte, 16*a.rows*a.cols)
	for i := range a.data {
		a.data[i].PutBytes(aa[16*i:])
	}

	oo := make([]byte, 16*a.rows*b.cols)

	C.binary_multiply128(C.int(a.rows), C.int(a.cols), C.int(b.cols),
		(*C.__uint128_t)((*[16]byte)(aa[:16])),
		(*C.uint8_t)(&b.data[0]),
		(*C.__uint128_t)((*[16]byte)(oo[:16])),
	)

	out := New128(a.rows, b.cols)
	for i := range out.data {
		out.data[i] = uint128.FromBytes(oo[i*16:])
	}

	return out
}

func Mul128(a *Matrix128, b *Matrix128) *Matrix128 {
	if a.cols != b.rows {
		panic("Dimension mismatch")
	}

	aa := make([]byte, 16*a.rows*a.cols)
	for i := range a.data {
		a.data[i].PutBytes(aa[16*i:])
	}

	bb := make([]byte, 16*b.rows*b.cols)
	for i := range b.data {
		b.data[i].PutBytes(bb[16*i:])
	}

	oo := make([]byte, 16*a.rows*b.cols)

	C.multiply128(C.int(a.rows), C.int(a.cols), C.int(b.cols),
		(*C.__uint128_t)((*[16]byte)(aa[:16])),
		(*C.__uint128_t)((*[16]byte)(bb[:16])),
		(*C.__uint128_t)((*[16]byte)(oo[:16])),
	)

	out := New128(a.rows, b.cols)
	for i := range out.data {
		out.data[i] = uint128.FromBytes(oo[i*16:])
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
