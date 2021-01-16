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
  var beta field.Element
  beta.SetUint64(7613)
	a, b := Gen(alpha, &beta, logN)

  var sum field.Element
  var out0, out1 field.Element
  zero := field.Zero()
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		Eval(a, i, logN, &out0)
		Eval(b, i, logN, &out1)

    sum.Add(&out0, &out1)

    //log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
    if i != alpha && !sum.Equal(&zero) {
      test.Fail()
    }

    if i == alpha && !sum.Equal(&beta) {
			test.Fail()
		}
	}
}

func TestEvalFull(test *testing.T) {
	logN := uint64(9)
	alpha := uint64(128)
  var beta field.Element
  beta.SetUint64(7613)
  outA := make([]field.Element, 1 << logN)
  outB := make([]field.Element, 1 << logN)

	a, b := Gen(alpha, &beta, logN)
	EvalFull(a, logN, outA)
	EvalFull(b, logN, outB)

  zero := field.Zero()
  var sum field.Element
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		sum.Add(&outA[i], &outB[i])
    if i != alpha && !sum.Equal(&zero) {
      test.Fail()
    }

    if i == alpha && !sum.Equal(&beta) {
			test.Fail()
		}
	}
}

func TestEvalFullShort(test *testing.T) {
	logN := uint64(3)
	alpha := uint64(2)
  var beta field.Element
  beta.SetUint64(7613)
  outA := make([]field.Element, 1 << logN)
  outB := make([]field.Element, 1 << logN)

	a, b := Gen(alpha, &beta, logN)
	EvalFull(a, logN, outA)
	EvalFull(b, logN, outB)

  zero := field.Zero()
  var sum field.Element
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		sum.Add(&outA[i], &outB[i])
    if i != alpha && !sum.Equal(&zero) {
      test.Fail()
    }

    if i == alpha && !sum.Equal(&beta) {
			test.Fail()
		}
	}
}
