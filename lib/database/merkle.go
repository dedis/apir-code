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

	for i, _ := range entriesFlatten {
		r := bytes.NewReader(entriesFlatten)
		merkleRoot, proof, numLeaves, err := merkletree.BuildReaderProof(r, sha256.New(), blockLen*field.Bytes, uint64(i))
		if err != nil {
			panic(err)
		}
		fmt.Println(merkleRoot, proof, numLeaves)
	}

	return nil
}

func flatten(m [][]byte) []byte {
	out := make([]byte, len(m)*len(m[0]))
	for i := range m {
		for j := range m[0] {
			out[i*len(m)+j] = m[i][j]
		}
	}

	return out
}
