package field

import (
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

func BenchmarkInnerProduct(b *testing.B) {
	rnd1 := utils.RandomPRG()
	rnd2 := utils.RandomPRG()

	length := 10000

	aa, err := RandomVectorPointers(rnd1, length)
	if err != nil {
		panic(err)
	}
	bb, err := RandomVectorPointers(rnd2, length)
	if err != nil {
		panic(err)
	}
	r := new(Element).SetZero()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i, a := range aa {
			r.Add(r, a.Mul(a, bb[i]))
		}
	}

}

func BenchmarkMul(b *testing.B) {
	rnd := utils.RandomPRG()
	x, err := new(Element).SetRandom(rnd)
	if err != nil {
		panic(err)
	}
	y, err := new(Element).SetRandom(rnd)
	if err != nil {
		panic(err)
	}
	r := new(Element).SetOne()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Mul(x, y)
	}
}