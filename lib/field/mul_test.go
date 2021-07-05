package field

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkMul32(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	length := 10000

	xx := make([]uint32, length)
	yy := make([]uint32, length)

	for i := range xx {
		xx[i] = uint32(rand.Int31())
		yy[i] = uint32(rand.Int31())
	}
	z := uint64(0)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range xx {
			z += uint64(xx[i]) * uint64(yy[i])
		}
	}
}
