package matrix

import (
	"encoding/binary"
	"io"
	"math"
	"math/rand"

	"github.com/si-co/vpir-code/lib/utils"
)

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

func NewRandom(rnd io.Reader, r int, c int, mod uint64) *Matrix {
	if mod != 1<<32 {
		panic("change function")
	}
	bytesMod := utils.ParamsDefault().Bytes
	data := make([]byte, bytesMod*r*c)
	if _, err := rnd.Read(data); err != nil {
		panic(err)
	}

	m := New(r, c)
	for i := 0; i < len(m.data); i++ {
		m.data[i] = binary.BigEndian.Uint32(data[i*bytesMod : (i+1)*bytesMod])
	}

	return m
}

func sampleGauss(sigma float64) uint32 {
	// TODO: Replace with cryptographic RNG

	// Inspired by https://github.com/malb/dgs/
	tau := float64(18)
	upper_bound := int(math.Ceil(sigma * tau))
	f := -1.0 / (2.0 * sigma * sigma)

	x := 0
	for {
		// Sample random value in [-tau*sigma, tau-sigma]
		x = rand.Intn(2*upper_bound+1) - upper_bound
		diff := float64(x)
		accept_with_prob := math.Exp(diff * diff * f)
		if rand.Float64() < accept_with_prob {
			break
		}
	}

	return uint32(x)
}

func NewGauss(r int, c int, sigma float64) *Matrix {
	m := New(r, c)
	for i := 0; i < len(m.data); i++ {
		m.data[i] = sampleGauss(sigma)
	}

	return m
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

func Mul(a *Matrix, b *Matrix) *Matrix {
	if a.cols != b.rows {
		panic("Dimension mismatch")
	}

	// TODO Implement this inner loop in C for performance
	out := New(a.rows, b.cols)
	for i := 0; i < a.rows; i++ {
		for k := 0; k < a.cols; k++ {
			for j := 0; j < b.cols; j++ {
				out.data[b.cols*i+j] += a.data[a.cols*i+k] * b.data[b.cols*k+j]
			}
		}
	}

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
