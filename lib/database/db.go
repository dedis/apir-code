package database

import (
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/field"
)

var text = "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"

type DB struct {
	Entries [][]field.Element
	Info
}

type Info struct {
	NumRows    int
	NumColumns int
	BlockSize  int

	// embedding info
	IDLength  int
	KeyLength int
}

func CreateZeroMultiBitDB(numRows, numColumns, blockSize int) *DB {
	entries := make([][]field.Element, numRows)
	for i := 0; i < numRows; i++ {
		entries[i] = field.ZeroVector(numColumns * blockSize)
	}
	return &DB{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockSize,
		},
	}
}

func CreateRandomMultiBitDB(rnd io.Reader, dbLen, numRows, blockLen int) *DB {
	var err error
	entries := make([][]field.Element, numRows)
	numColumns := dbLen / (128 * numRows * blockLen)
	for i := 0; i < numRows; i++ {
		entries[i], err = field.RandomVector(rnd, numColumns*blockLen)
		if err != nil {
			log.Fatal(err)
		}
	}
	return &DB{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
		},
	}
}

func CreateRandomSingleBitDB(rnd io.Reader, dbLen, numRows int) *DB {
	var tmp field.Element
	entries := make([][]field.Element, numRows)
	numColumns := dbLen / numRows
	for i := 0; i < numRows; i++ {
		entries[i] = make([]field.Element, numColumns)
		for j := 0; j < numColumns; j++ {
			tmp.SetRandom(rnd)
			tmpb := tmp.Bytes()[len(tmp.Bytes())-1]
			if tmpb>>7 == 1 {
				entries[i][j].SetOne()
			} else {
				entries[i][j].SetZero()
			}
		}
	}

	return &DB{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: 0,
		},
	}
}
