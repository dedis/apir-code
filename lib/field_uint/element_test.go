package field

import (
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

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
