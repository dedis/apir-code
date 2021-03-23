package database

import (
	"io"
	"log"
)

type Bytes struct {
	Entries []byte
	Info
}

// blockLen must be the number of bytes in a block, as a byte is the element
func CreateRandomMultiBitBytes(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	entries := make([]byte, dbLen/8)
	numColumns := dbLen / (8 * numRows * blockLen)
	if _, err := rnd.Read(entries); err != nil {
		log.Fatal(err)
	}
	return &Bytes{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
		},
	}
}
