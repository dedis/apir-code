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
		copy(blocks[i], randomBlocks[i*blockLen:(i+1)*blockLen])
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}

	// generate db
	numColumns := numBlocks / numRows
	proofLen := tree.EncodedProofLength()
	blockLen = blockLen + proofLen
	entries := make([]byte, numRows*numColumns*blockLen)
	b := 0
	for i := 0; i < numRows*numColumns*blockLen; i += blockLen {
		p, err := tree.GenerateProof(blocks[b])
		if err != nil {
			log.Fatalf("error while generating proof for block %v: %v", b, err)
		}
		encodedProof := merkle.EncodeProof(p)
		copy(entries[i:i+blockLen], append(blocks[b], encodedProof...))
		b++
	}

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: numColumns,
			BlockSize:  blockLen,
			PIRType:    "merkle",
			Root:       tree.Root(),
			ProofLen:   proofLen,
		},
	}

	return m
}
