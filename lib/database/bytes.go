package database

import (
	"io"
	"log"
)

type Bytes struct {
	Entries []byte
	Info
}

// CreateBitBytes return a random bytes database.
// blockLen must be the number of bytes in a block, as a byte is the element
func CreateZeroBytes(numRows, numColumns, blockLen int) *Bytes {
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

// InitBytes return an empty database with a initial zero capacity, to
// be used when embedding keys into a bytes database.
// blockLen must be the number of bytes in a block, as a byte is the element
func InitBytes(numRows, numColumns, blockLen int) *Bytes {
	// sample random entries
	entries := make([]byte, 0)

	return &Bytes{Entries: entries,
		Info: Info{
			NumColumns:   numColumns,
			NumRows:      numRows,
			BlockSize:    blockLen,
			BlockLengths: make([]int, numRows*numColumns),
			Merkle:       &Merkle{ProofLen: 0}, // only for tests compatibility
		},
	}
}

// CreateRandomBytes return a random bytes database.
// blockLen must be the number of bytes in a block, as a byte is the element
func CreateRandomBytes(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	// sample random entries
	entries := make([]byte, dbLen/8)
	if _, err := rnd.Read(entries); err != nil {
		log.Fatal(err)
	}

	numColumns := dbLen / (8 * numRows * blockLen)
	blockLens := make([]int, numRows*numColumns)
	for i := 0; i < numRows*numColumns; i++ {
		blockLens[i] = blockLen
	}
	return &Bytes{Entries: entries,
		Info: Info{
			NumColumns:   numColumns,
			NumRows:      numRows,
			BlockSize:    blockLen,
			BlockLengths: blockLens,
			Merkle:       &Merkle{ProofLen: 0}, // only for tests compatibility
		},
	}
}

func (b *Bytes) SizeGiB() float64 {
	return float64(len(b.Entries)) * 9.313e-10
}
