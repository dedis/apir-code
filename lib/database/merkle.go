package database

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/si-co/vpir-code/lib/field"

	"gitlab.com/NebulousLabs/merkletree"
)

type Merkle struct {
	db *Bytes
}

// CreateRandomMultiBitMerkle
// blockLen is the number of byte in a block, as byte is viewd as an element in this
// case
func CreateRandomMultiBitMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Merkle {
	db := CreateRandomMultiBitBytes(rnd, dbLen, numRows, blockLen)
	entriesFlatten := flatten(db.Entries)
	fmt.Println(entriesFlatten)
	r := bytes.NewReader(entriesFlatten)

	for i, _ := range entriesFlatten {
		merkleRoot, proof, numLeaves, _ := merkletree.BuildReaderProof(r, sha256.New(), blockLen*field.Bytes, uint64(i))
		fmt.Println(merkleRoot, proof, numLeaves)
	}

	return nil
}

func flatten(m [][]byte) []byte {
	return m[0][:cap(m[0])]
}
