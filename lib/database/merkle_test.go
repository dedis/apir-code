package database

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	merkletree "github.com/wealdtech/go-merkletree"
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
	tree, err := merkletree.New(blocks)
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
			fmt.Printf("Original proof: %#v\n", p)
			require.NoError(t, err)
			encodedProof := encodeProof(p)
			e = append(e, append(blocks[b], encodedProof...)...)
			proofLen = len(encodedProof) // always same length

			//fmt.Println("Original:", hex.EncodeToString(blocks[b]), hex.EncodeToString(encodedProof))
			fmt.Println("")

			// first verification here
			verified, err := merkletree.VerifyProof(blocks[b], p, root)
			require.NoError(t, err)
			require.True(t, verified)

			b++
			if j == 1 {
				break
			}
		}
		entries[i] = e
	}

	// verify db
	for i := range entries {
		for j := 0; j < blocksPerRow-1; j++ {
			entireBlock := entries[i][j*(blockLen+proofLen) : (j+1)*(blockLen+proofLen)]
			data := entireBlock[:blockLen]
			encodedProof := entireBlock[blockLen:]
			//fmt.Println("Extracted:", hex.EncodeToString(data), hex.EncodeToString(encodedProof))
			fmt.Println("")
			proof := DecodeProof(encodedProof)
			fmt.Printf("Extracted proof: %#v\n", proof)
			verified, err := merkletree.VerifyProof(data, proof, root)
			require.NoError(t, err)
			require.True(t, verified)
			fmt.Println(verified)
			if j == 1 {
				break
			}
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
	tree, err := merkletree.New(data)
	require.NoError(t, err)

	// generate a proof for random element
	proof, err := tree.GenerateProof(data[rand.Intn(len(data))])
	require.NoError(t, err)

	// encode the proof
	b := encodeProof(proof)

	// decode proof
	p := DecodeProof(b)

	require.Equal(t, *proof, *p)
}
