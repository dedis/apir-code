
package matrix

import (
  crand "crypto/rand"
  "math"
  "math/rand"
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
    v, err := crand.Int(crand.Reader, modBig)
    if err != nil {
      panic("Error reading random int")
    }
    m.data[i] = uint32(v.Uint64())
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
  for ;; {
    // Sample random value in [-tau*sigma, tau-sigma]
    x := rand.Intn(2*upper_bound + 1) - upper_bound

    diff := float64(x)
    accept_with_prob := math.Exp(diff * diff * f)
    if rand.Float64() >= accept_with_prob {
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
  m.data[m.cols * r + c] = v
}

func (m *Matrix) Get(r int, c int) uint32 {
  return m.data[m.cols * r + c]
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

func (a *Matrix) Add(b *Matrix) {
  if a.cols != b.cols || a.rows != b.rows{
    panic("Dimension mismatch")
  }

  for i := 0; i < len(a.data); i++ {
    a.data[i] += b.data[i]
  }
}

func (a *Matrix) Sub(b *Matrix) {
  if a.cols != b.cols || a.rows != b.rows{
    panic("Dimension mismatch")
  }

  for i := 0; i < len(a.data); i++ {
    a.data[i] -= b.data[i]
  }
}
