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
	randomBlocks := make([]byte, numBlocks*blockLen)
	if _, err := rnd.Read(randomBlocks); err != nil {
		log.Fatal(err)
	}

	blocks := make([][]byte, numBlocks)
	for i := range blocks {
		// generate random block
		blocks[i] = make([]byte, blockLen)
		copy(blocks[i], randomBlocks[i*blockLen : (i+1)*blockLen])
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}

	// generate db
	blocksPerRow := numBlocks / numRows
	proofLen := tree.EncodedProofLength()
	columnLen := blockLen + proofLen
	b := 0
	for i := range entries {
		e := make([]byte, columnLen * blocksPerRow)
		for j := 0; j < blocksPerRow; j++ {
			p, err := tree.GenerateProof(blocks[b])
			encodedProof := merkle.EncodeProof(p)
			if err != nil {
				log.Fatalf("error while generating proof for block %v: %v", b, err)
			}
			copy(e[j*columnLen:(j+1)*columnLen], append(blocks[b], encodedProof...))
			b++
		}
		entries[i] = e
	}

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: dbLen / (8 * numRows * blockLen),
			BlockSize:  columnLen,
			PIRType:    "merkle",
			Root:       tree.Root(),
			ProofLen:   proofLen,
		},
	}

	return m
}
