package matrix

import (
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/si-co/vpir-code/lib/utils"
)

/*
#cgo CFLAGS: -std=c99 -O3 -march=native -msse4.1 -maes -mavx2 -mavx
#include <matrix.h>
*/
import "C"

type Matrix struct {
	rows int
	cols int
	data []uint32
}

func New(r int, c int) *Matrix {
	return &Matrix{
		rows: r,
		cols: c,
		data: make([]uint32, r*c),
	}
}

func NewWithData(r int, c int, data []uint32) *Matrix {
	return &Matrix{
		rows: r,
		cols: c,
		data: data,
	}
}

func MatrixToBytes(in *Matrix) []byte {
	// we first store rows and cols to allow reconstruction
	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, uint32(in.rows))
	c := make([]byte, 4)
	binary.BigEndian.PutUint32(c, uint32(in.cols))
	params := append(r, c...)
	// finally we store the data and append the params in front of them
	return append(params, utils.Uint32SliceToByteSlice(in.data)...)
}

func BytesToMatrix(in []byte) *Matrix {
	// retrieve the matrix dimensions
	r := in[:4]
	rows := int(binary.BigEndian.Uint32(r))
	c := in[4:8]
	cols := int(binary.BigEndian.Uint32(c))
	data := utils.ByteSliceToUint32Slice(in[8:])
	return &Matrix{
		rows: rows,
		cols: cols,
		data: data,
	}
}

func MatricesToBytes(in []*Matrix) []byte {
	// the matrices are all the same, so
	// we can use MatrixToBytes to see how many
	// bytes are necessary for one, and then encode the length
	// and all the matrices
	b := MatrixToBytes(in[0])

	out := make([]byte, 4+len(b)*len(in))
	binary.BigEndian.PutUint32(out[:4], uint32(len(b)))
	copy(out[4:], b)

	// first matrix is already encoded
	for i := 1; i < len(in); i++ {
		if in[0].rows != in[i].rows || in[0].cols != in[i].cols {
			panic("dimension mismatch")
		}
		copy(out[4+len(b)*i:], MatrixToBytes(in[i]))
	}

	return out
}

func BytesToMatrices(in []byte) []*Matrix {
	// retrieve length of single encoded matrix
	length := int(binary.BigEndian.Uint32(in[:4]))

	// how many matrices?
	n := uint32((len(in) - 4) / length)

	out := make([]*Matrix, n)

	for i := range out {
		out[i] = BytesToMatrix(in[4+i*length : 4+(i+1)*length])
	}

	return out
}

func NewRandom(rnd io.Reader, r int, c int) *Matrix {
	bytesMod := utils.ParamsDefault().BytesMod
	b := make([]byte, bytesMod*r*c)
	if _, err := rnd.Read(b); err != nil {
		panic(err)
	}

	m := New(r, c)

	// ugly hack for performance
	m.data = *(*[]uint32)(unsafe.Pointer(&b))

	return m
}

func NewGauss(r int, c int) *Matrix {
	m := New(r, c)
	for i := 0; i < len(m.data); i++ {
		m.data[i] = uint32(utils.GaussSample())
	}

	return m
}

func (m *Matrix) SetData(i int, v uint32) {
	m.data[i] = v
}

func (m *Matrix) Set(r int, c int, v uint32) {
	m.data[m.cols*r+c] = v
}

func (m *Matrix) Get(r int, c int) uint32 {
	return m.data[m.cols*r+c]
}

func (m *Matrix) Rows() int {
	return m.rows
}

func (m *Matrix) Cols() int {
	return m.cols
}

func (m *Matrix) Len() int {
	return len(m.data)
}

func BinaryMul(a *Matrix, b *MatrixBytes) *Matrix {
	if a.cols != b.rows {
		panic("Dimension mismatch")
	}

	out := New(a.rows, b.cols)
	C.binary_multiply(C.int(a.rows), C.int(a.cols), C.int(b.cols),
		(*C.uint32_t)(&a.data[0]), (*C.uint8_t)(&b.data[0]),
		(*C.uint32_t)(&out.data[0]))

	// for i := 0; i < a.rows; i++ {
	// 	for k := 0; k < a.cols; k++ {
	// 		for j := 0; j < b.cols; j++ {
	// 			if b.data[b.cols*k+j] != byte(0) {
	// 				out.data[b.cols*i+j] += a.data[a.cols*i+k]
	// 			}
	// 		}
	// 	}
	// }

	return out
}

func Mul(a *Matrix, b *Matrix) *Matrix {
	if a.cols != b.rows {
		panic("Dimension mismatch")
	}

	out := New(a.rows, b.cols)
	C.multiply(C.int(a.rows), C.int(a.cols), C.int(b.cols),
		(*C.uint32_t)(&a.data[0]), (*C.uint32_t)(&b.data[0]),
		(*C.uint32_t)(&out.data[0]))

	return out
}

func (a *Matrix) Add(b *Matrix) {
	if a.cols != b.cols || a.rows != b.rows {
		panic("Dimension mismatch")
	}

	for i := 0; i < len(a.data); i++ {
		a.data[i] += b.data[i]
	}
}

func (a *Matrix) Sub(b *Matrix) {
	if a.cols != b.cols || a.rows != b.rows {
		panic("Dimension mismatch")
	}

	for i := 0; i < len(a.data); i++ {
		a.data[i] -= b.data[i]
	}
}

func (a *Matrix) BytesSize() float64 {
	return float64(a.rows * a.cols * 4)
}
