
package main

import "fmt"
import "github.com/si-co/lwe/matrix"

type Params struct {
  p uint32    // plaintext modulus
  n int       // lattice/secret dimension

  l int       // number of rows of database
  m int       // number of columns of database

  A *matrix.Matrix   // Matrix used to generate digest
}

func matrixDB(p *Params, db []uint32) *matrix.Matrix {
  if len(db) != p.m * p.l {
    panic("Invalid database dimensions")
  }

  out := matrix.New(p.l, p.m)
  for i := 0; i < p.l; i++ {
    for j := 0; j < p.m; j++ {
      val := db[i*p.m + j]
      if val >= p.p {
        panic("Plaintext value too large")
      }
      out.Set(i, j, val)
    }
  }

  return out
}

func Digest(p *Params, db *matrix.Matrix) *matrix.Matrix {
  return matrix.Mul(p.A, db)
}

func ParamsDefault() *Params {
  p := &Params{
    p: 10 ,
    n: 1024,
    l: 512,
    m: 512,
  }

  p.A = matrix.NewRandom(p.n, p.l, 1 << 32)
  return p
}

func RandomDB(p *Params) *matrix.Matrix {
  data := make([]uint32, p.l * p.m)
  for i := 0; i < len(data); i++ {
    // TODO: Replace with something real
    data[i] = uint32(i % 2)
  }
  return matrixDB(p, data)
}

func main() {
  p := ParamsDefault()
  db := RandomDB(p)
  d := Digest(p, db)

  fmt.Printf("%v x %v\n", d.Rows(), d.Cols())
}
