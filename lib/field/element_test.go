package field

import (
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

func Benchmark256Bits(b *testing.B) {
	rnd := utils.RandomPRG()

	// 8 elements of 32 bits each
	elements := make([]*Element, 8)
	var err error
	for i := range elements {
		elements[i], err = new(Element).SetRandom(rnd)
		if err != nil {
			panic(err)
		}
	}

	r := new(Element).SetOne()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range elements {
			r.Mul(r, e)
		}
	}

}

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
	r64 := uint64(0)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i, a := range aa {
			prod := a.Mul(a, bb[i])
			r64 += uint64(prod.E)
		}

		r64 %= uint64(p)
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
