package database

import (
	"io"
	"log"
)

type Bytes struct {
	Entries [][]byte
	Info
}

func CreateZeroMultiBitBytes(numRows, numColumns, blockSize int) *Bytes {
	entries := make([][]byte, numRows)
	for i := 0; i < numRows; i++ {
		// default value for []byte is already zero
		entries[i] = make([]byte, numColumns*blockSize)
	}
	return &Bytes{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockSize,
		},
	}
}

func CreateRandomMultiBitBytes(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	entries := make([][]byte, numRows)
	numColumns := dbLen / (128 * numRows * blockLen)
	for i := 0; i < numRows; i++ {
		e := make([]byte, numColumns*blockLen)
		if _, err := rnd.Read(e); err != nil {
			log.Fatal(err)
		}
		entries[i] = e
	}
	return &Bytes{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
		},
	}
}
