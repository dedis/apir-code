
package matrix

import (
  "crypto/rand"
  "math/big"
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

func NewRandom(r int, c int, mod uint64) *Matrix {
  modBig := big.NewInt(int64(mod))

  // TODO: Replace with something much faster
  m := New(r, c)
  for i := 0; i < len(m.data); i++ {
    v, err := rand.Int(rand.Reader, modBig)
    if err != nil {
      panic("Error reading random int")
    }
    m.data[i] = uint32(v.Uint64())
  }

  return m
}

func (m *Matrix) Set(r int, c int, v uint32) {
  m.data[m.cols * r + c] = v
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
        out.data[b.cols*i + j] += a.data[a.cols*i + k] * b.data[b.cols*k + j]
      }
    }
  }

  return out
}

