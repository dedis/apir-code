package database

import (
	"log"
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/merkle"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

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
			p, err := tree.GenerateProof(blocks[b], 0)
			require.NoError(t, err)
			encodedProof := encodeProof(p)
			e = append(e, append(blocks[b], encodedProof...)...)
			proofLen = len(encodedProof) // always same length

			// first verification here
			verified, err := merkle.VerifyProof(blocks[b], false, p, [][]byte{root})
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
			proof := DecodeProof(encodedProof)
			verified, err := merkle.VerifyProof(data, false, proof, [][]byte{root})
			require.NoError(t, err)
			require.True(t, verified)
		}
	}
}

func TestEncodeDecodeProof(t *testing.T) {
	rng := utils.RandomPRG()
	data := make([][]byte, rand.Intn(501))
	for i := range data {
		d := make([]byte, 32)
		rng.Read(d)
		data[i] = d
	}

	// create the tree
	tree, err := merkle.New(data)
	require.NoError(t, err)

	// generate a proof for random element
	proof, err := tree.GenerateProof(data[rand.Intn(len(data))], 0)
	require.NoError(t, err)

	// encode the proof
	b := encodeProof(proof)

	// decode proof
	p := DecodeProof(b)

	require.Equal(t, *proof, *p)
}
