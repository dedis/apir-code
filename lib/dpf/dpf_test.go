package dpf

import (
//  "log"
	"testing"

	"github.com/si-co/vpir-code/lib/field"
)

/*
func BenchmarkEvalFull(bench *testing.B) {
	logN := uint64(28)
	a, _ := Gen(0, logN)
	bench.ResetTimer()
	//fmt.Println("Ka: ", a)
	//fmt.Println("Kb: ", b)
	//for i:= uint64(0); i < (uint64(1) << logN); i++ {
	//	aa := dpf.Eval(a, i, logN)
	//	bb := dpf.Eval(b, i, logN)
	//	fmt.Println(i,"\t", aa,bb, aa^bb)
	//}
	for i := 0; i < bench.N; i++ {
		EvalFull(a, logN)
	}
}
*/

func BenchmarkXor16(bench *testing.B) {
	a := new(block)
	b := new(block)
	c := new(block)
	for i := 0; i < bench.N; i++ {
		xor16(&c[0], &b[0], &a[0])
	}
}

func TestEval(test *testing.T) {
	logN := uint64(8)
	alpha := uint64(123)
  beta := make([]field.Element, 2)
  beta[0].SetUint64(7613)
  beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

  sum := make([]field.Element, 2)
  out0 := make([]field.Element, 2)
  out1 := make([]field.Element, 2)
  zero := field.Zero()
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		Eval(a, i, logN, out0)
		Eval(b, i, logN, out1)

    for j := 0; j < 2; j++ {
      sum[j].Add(&out0[j], &out1[j])
    }

    //log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
    if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
      test.Fail()
    }

    if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}

func TestEvalFull(test *testing.T) {
	logN := uint64(9)
	alpha := uint64(123)
  beta := make([]field.Element, 2)
  beta[0].SetUint64(7613)
  beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

  sum := make([]field.Element, 2)
  out0 := make([][]field.Element, 1 << logN)
  out1 := make([][]field.Element, 1 << logN)

  for i := 0; i < len(out0); i++ {
    out0[i] = make([]field.Element, 2)
    out1[i] = make([]field.Element, 2)
  }

	EvalFull(a, logN, out0)
	EvalFull(b, logN, out1)

  zero := field.Zero()
	for i := uint64(0); i < (uint64(1) << logN); i++ {
    for j := 0; j < 2; j++ {
      sum[j].Add(&out0[i][j], &out1[i][j])
    }

    //log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
    if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
      test.Fail()
    }

    if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}

func TestEvalFullShort(test *testing.T) {
	logN := uint64(2)
	alpha := uint64(2)
  beta := make([]field.Element, 2)
  beta[0].SetUint64(7613)
  beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

  sum := make([]field.Element, 2)
  out0 := make([][]field.Element, 1 << logN)
  out1 := make([][]field.Element, 1 << logN)

  for i := 0; i < len(out0); i++ {
    out0[i] = make([]field.Element, 2)
    out1[i] = make([]field.Element, 2)
  }

	EvalFull(a, logN, out0)
	EvalFull(b, logN, out1)

  zero := field.Zero()
	for i := uint64(0); i < (uint64(1) << logN); i++ {
    for j := 0; j < 2; j++ {
      sum[j].Add(&out0[i][j], &out1[i][j])
    }

    //log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
    if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
      test.Fail()
    }

    if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}

func TestEvalFullPartial(test *testing.T) {
	logN := uint64(9)
	alpha := uint64(123)
  beta := make([]field.Element, 2)
  beta[0].SetUint64(7613)
  beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

  sum := make([]field.Element, 2)

  outlen := 278
  out0 := make([][]field.Element, outlen)
  out1 := make([][]field.Element, outlen)

  for i := 0; i < len(out0); i++ {
    out0[i] = make([]field.Element, 2)
    out1[i] = make([]field.Element, 2)
  }

	EvalFull(a, logN, out0)
	EvalFull(b, logN, out1)

  zero := field.Zero()
	for i := uint64(0); i < uint64(outlen); i++ {
    for j := 0; j < 2; j++ {
      sum[j].Add(&out0[i][j], &out1[i][j])
    }

    //log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
    if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
      test.Fail()
    }

    if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}
