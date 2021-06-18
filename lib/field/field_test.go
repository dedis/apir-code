package field

import (
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

func Benchmark256Bits(b *testing.B) {
	rnd := utils.RandomPRG()

	// 2 elements of 32 bits each
	elements := make([]*Element, 2)
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
	r := new(Element).SetZero()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i, a := range aa {
			r.Add(r, a.Mul(a, bb[i]))
		}
	}

}

func BenchmarkSetFixedBytes(b *testing.B) {
	rng := utils.RandomPRG()
	var buf [16]byte
	rng.Read(buf[:])

	z := new(Element)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		z.SetFixedLengthBytes(buf)
	}
}
