package merkle

import (
	"log"
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	rnd := utils.RandomPRG()

	// generate random blocks
	blockLen := 256
	blocks := make([][]byte, 1000)
	for i := range blocks {
		// generate random block
		b := make([]byte, blockLen)
		if _, err := rnd.Read(b); err != nil {
			log.Fatal(err)
		}
		blocks[i] = b
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// generate tree
		_, err := New(blocks)
		if err != nil {
			panic(err)
		}
	}
}
