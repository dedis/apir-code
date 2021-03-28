package field

import (
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

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
