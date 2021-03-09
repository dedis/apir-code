package database

import (
	"log"
	"testing"

	"github.com/si-co/vpir-code/lib/merkle"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func BenchmarkMerkleTree(b *testing.B) {
	b.ReportAllocs()
	rng := utils.RandomPRG()
	dbLen := 100000
	numRows := 1
	blockLen := 160

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateRandomMultiBitMerkle(rng, dbLen, numRows, blockLen)
	}
}

func TestMerkleTree(t *testing.T) {
	rng := utils.RandomPRG()
	dbLen := 100000
	numRows := 1
	blockLen := 160

	entries := make([][]byte, numRows)
	numBlocks := dbLen / (8 * blockLen)
	// generate random blocks
	blocks := make([][]byte, numBlocks)
	for i := range blocks {
		// generate random block
		b := make([]byte, blockLen)
		if _, err := rng.Read(b); err != nil {
			log.Fatal(err)
		}
		blocks[i] = b
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}
	root := tree.Root()

	// generate db
	blocksPerRow := numBlocks / numRows
	proofLen := 0
	b := 0
	for i := range entries {
		e := make([]byte, 0)
		for j := 0; j < blocksPerRow; j++ {
			p, err := tree.GenerateProof(blocks[b])
			require.NoError(t, err)
			encodedProof := merkle.EncodeProof(p)
			e = append(e, append(blocks[b], encodedProof...)...)
			proofLen = len(encodedProof) // always same length

			// first verification here
			verified, err := merkle.VerifyProof(blocks[b], false, p, root)
			require.NoError(t, err)
			require.True(t, verified)

			b++
		}
		entries[i] = e
	}

	// verify db
	for i := range entries {
		for j := 0; j < blocksPerRow-1; j++ {
			entireBlock := entries[i][j*(blockLen+proofLen) : (j+1)*(blockLen+proofLen)]
			data := entireBlock[:blockLen]
			encodedProof := entireBlock[blockLen:]
			proof := merkle.DecodeProof(encodedProof)
			verified, err := merkle.VerifyProof(data, false, proof, root)
			require.NoError(t, err)
			require.True(t, verified)
		}
	}
}
