package database

import (
	"bytes"
	"io"
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
	r := bytes.NewReader(byteData)

	return nil
}

func flatten(m [][]byte) []byte {
	return m[0][:cap(m[0])]
}
