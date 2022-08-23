package matrix

import (
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/si-co/vpir-code/lib/utils"
)

/*
#cgo CFLAGS: -std=c99 -O3
#include <stdint.h>

void multiply(int aRows, int aCols, int bCols, uint64_t *a, uint64_t *b, uint64_t *out) {
   	int i, j, k;
	for (i = 0; i < aRows; i++) {
		for (k = 0; k < aCols; k++) {
			for (j = 0; j < bCols; j++) {
				out[bCols*i+j] += a[aCols*i+k] * b[bCols*k+j];
			}
		}
	}
}
*/
import "C"

type Matrix struct {
	rows int
	cols int
	data []uint64
}

func New(r int, c int) *Matrix {
	return &Matrix{
		rows: r,
		cols: c,
		data: make([]uint64, r*c),
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
	return append(params, utils.Uint64SliceToByteSlice(in.data)...)
}

func BytesToMatrix(in []byte) *Matrix {
	// retrieve the matrix dimensions
	r := in[:4]
	rows := int(binary.BigEndian.Uint32(r))
	c := in[4:8]
	cols := int(binary.BigEndian.Uint32(c))
	data := utils.ByteSliceToUint64Slice(in[8:])
	return &Matrix{
		rows: rows,
		cols: cols,
		data: data,
	}
}

func NewRandom(rnd io.Reader, r int, c int) *Matrix {
	bytesMod := utils.ParamsDefault().BytesMod
	b := make([]byte, bytesMod*r*c)
	if _, err := rnd.Read(b); err != nil {
		panic(err)
	}

	m := New(r, c)

	// for i := 0; i < len(m.data); i++ {
	// 	m.data[i] = binary.BigEndian.Uint32(b[i*bytesMod : (i+1)*bytesMod])
	// }
	// TODO: this works but it is bad practice
	m.data = *(*[]uint64)(unsafe.Pointer(&b))

	return m
}

func NewGauss(r int, c int, sigma float64) *Matrix {
	m := New(r, c)
	for i := 0; i < len(m.data); i++ {
		m.data[i] = uint64(utils.GaussSample())
	}

	return m
}

func (m *Matrix) Set(r int, c int, v uint64) {
	m.data[m.cols*r+c] = v
}

func (m *Matrix) Get(r int, c int) uint64 {
	return m.data[m.cols*r+c]
}

func (m *Matrix) Rows() int {
	return m.rows
}

func (m *Matrix) Cols() int {
	return m.cols
}

func Mul(a *Matrix, b *Matrix) *Matrix {
	if a.cols != b.rows {
		panic("Dimension mismatch")
	}

	// TODO: the overhead of the cgo call is quite high (more than 50% of
	// total execution time for small matrices), we might want to
	// replace this with native assmebly code in Go
	out := New(a.rows, b.cols)
	C.multiply(C.int(a.rows), C.int(a.cols), C.int(b.cols),
		(*C.uint64_t)(&a.data[0]), (*C.uint64_t)(&b.data[0]),
		(*C.uint64_t)(&out.data[0]))

	return out

	// out := New(a.rows, b.cols)
	// for i := 0; i < a.rows; i++ {
	// 	for k := 0; k < a.cols; k++ {
	// 		for j := 0; j < b.cols; j++ {
	// 			out.data[b.cols*i+j] += a.data[a.cols*i+k] * b.data[b.cols*k+j]
	// 		}
	// 	}
	// }

	// return out
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
