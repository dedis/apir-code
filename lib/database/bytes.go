package database

import (
	"io"
	"log"
)

type Bytes struct {
	Entries []byte
	Info
}

// CreateRandomMultiBitBytes return a random bytes database.
// blockLen must be the number of bytes in a block, as a byte is the element
func CreateZeroMultiBitBytes(numRows, numColumns, blockLen int) *Bytes {
	// sample random entries
	entries := make([]byte, numRows*numColumns*blockLen)

	return &Bytes{Entries: entries,
		Info: Info{
			NumColumns: numColumns,
			NumRows:    numRows,
			BlockSize:  blockLen,
			Merkle:     &Merkle{ProofLen: 0}, // only for tests compatibility
		},
	}
}

// CreateRandomMultiBitBytes return a random bytes database.
// blockLen must be the number of bytes in a block, as a byte is the element
func CreateRandomMultiBitBytes(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	// sample random entries
	entries := make([]byte, dbLen/8)
	if _, err := rnd.Read(entries); err != nil {
		log.Fatal(err)
	}

	numColumns := dbLen / (8 * numRows * blockLen)
	return &Bytes{Entries: entries,
		Info: Info{
			NumColumns: numColumns,
			NumRows:    numRows,
			BlockSize:  blockLen,
			Merkle:     &Merkle{ProofLen: 0}, // only for tests compatibility
		},
	}
}
