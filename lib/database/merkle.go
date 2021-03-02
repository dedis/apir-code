package database

import (
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/merkle"
)

// CreateRandomMultiBitMerkle
// blockLen is the number of byte in a block, as byte is viewd as an element in this
// case
func CreateRandomMultiBitMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	entries := make([][]byte, numRows)
	numBlocks := dbLen / (8 * blockLen)
	// generate random blocks
	blocks := make([][]byte, numBlocks)
	for i := range blocks {
		// generate random block
		b := make([]byte, blockLen)
		if _, err := rnd.Read(b); err != nil {
			log.Fatal(err)
		}
		blocks[i] = b
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}

	// generate db
	blocksPerRow := numBlocks / numRows
	proofLen := 0
	b := 0
	for i := range entries {
		e := make([]byte, 0)
		for j := 0; j < blocksPerRow; j++ {
			p, err := tree.GenerateProof(blocks[b])
			encodedProof := merkle.EncodeProof(p)
			if err != nil {
				log.Fatalf("error while generating proof for block %v: %v", b, err)
			}
			e = append(e, append(blocks[b], encodedProof...)...)
			proofLen = len(encodedProof) // always same length
			b++
		}
		entries[i] = e
	}
	root := tree.Root()

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: dbLen / (8 * numRows * blockLen),
			BlockSize:  blockLen + proofLen,
			PIRType:    "merkle",
			Root:       root,
			ProofLen:   proofLen,
		},
	}

	return m
}
