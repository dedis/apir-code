
package main

import (
//  "fmt"
//  "log"
)
import "github.com/dedis/lwe/matrix"

// Ciphertext modulus
const MOD = 1 << 32

type Params struct {
  p uint32    // plaintext modulus
  n int       // lattice/secret dimension
  sigma float64 // Error parameter

  l int       // number of rows of database
  m int       // number of columns of database
  B uint32    // bound used in reconstruction

  A *matrix.Matrix   // Matrix used to generate digest
}

func RandomDB(p *Params) *matrix.Matrix {
  out := matrix.New(p.l, p.m)
  for i := 0; i < p.l; i++ {
    for j := 0; j < p.m; j++ {
      // TODO: Replace with something real
      val := uint32(3*uint32(i) + 7*uint32(j)) % p.p
      if val >= p.p {
        panic("Plaintext value too large")
      }
      out.Set(i, j, val)
    }
  }

  return out
}

func ParamsDefault() *Params {
  p := &Params{
    p: 2,
    n: 1024,
    sigma: 6.0,
    l: 512,
    m: 128,
    B: 1000,
  }

  p.A = matrix.NewRandom(p.n, p.l, MOD)
  return p
}

func Digest(p *Params, db *matrix.Matrix) *matrix.Matrix {
  // Digest has dimension n by m
  return matrix.Mul(p.A, db)
}

type State struct {
  digest *matrix.Matrix
  secret *matrix.Matrix
  i int
  j int
  t uint32
}

func Query(p *Params, digest *matrix.Matrix, i int, j int) (*State, *matrix.Matrix) {

  // Lazy way to sample a random scalar
  rand := matrix.NewRandom(1, 1, MOD)

  state := &State{
    digest: digest,
    secret: matrix.NewRandom(1, p.n, MOD),
    i: i,
    j: j,
    t: rand.Get(0, 0),
  }

  // Query has dimension 1 x l
  query := matrix.Mul(state.secret, p.A)

  // Error has dimension 1 x l
  e := matrix.NewGauss(1, p.l, p.sigma)

  msg := matrix.New(1, p.l)
  msg.Set(0, i, state.t)

  query.Add(e)
  query.Add(msg)

  return state, query
}

func Answer(p *Params, db *matrix.Matrix, query *matrix.Matrix) *matrix.Matrix {
  // Answer has dimension 1 x m
  return matrix.Mul(query, db)
}

func inRange(p *Params, val uint32) bool {
  return (val <= p.B) || (val >= -p.B)
}

func Reconstruct(p *Params, st *State, ans *matrix.Matrix) uint32 {
  s_trans_d := matrix.Mul(st.secret, st.digest)
  ans.Sub(s_trans_d)

  good := true
  outs := make([]uint32, p.m)
  for i := 0; i < p.m; i++ {
    v := ans.Get(0, i)
    //log.Printf("%v %v %v %v", v, v - st.t, p.B, -p.B)
    if inRange(p, v) {
      outs[i] = 0
    } else if inRange(p, v - st.t) {
      outs[i] = 1
    } else {
      good = false
    }
  }

  if !good {
    panic("Incorrect reconstruction")
  }

  return outs[st.j]
}

func main() {

  /*
  // Code that prints out the error distribution
  vals := make(map[uint32]int)
  for i := uint32(0); i < 30; i++ {
    vals[i] = 0
  }
  e := matrix.NewGauss(1, 10000, 3.0)
  for i := 0; i < e.Cols(); i++ {
    vals[e.Get(0,i)] += 1
  }

  for i := uint32(0); i < 30; i++ {
    v := uint32(0)
    log.Printf("%v = %v %v\n", i, vals[v-uint32(i)], vals[i])
  }
  */

  i,j := 7, 12

  p := ParamsDefault()
  db := RandomDB(p)
  digest := Digest(p, db)
  st,query := Query(p, digest, i, j)
  ans := Answer(p, db, query)
  out := Reconstruct(p, st, ans)

  if out != db.Get(i,j) {
    panic("Invalid reconstruction")
  }

}

