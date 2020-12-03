package database

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"math/big"
	"strconv"
)

type Database struct {
	Entries []*big.Int
}

func CreateDatabase() *Database {
	entries := make([]*big.Int, cst.DBLength)
	for i := 0; i < cst.DBLength; i++ {
		entries[i] = cst.BigOne
	}
	entries[9] = cst.BigZero

	return &Database{Entries: entries}
}

func CreateAsciiDatabase() *Database {
	// playing with VPIR in ascii
	text := "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"
	db := CreateDatabase()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		db.Entries[i] = new(big.Int).SetInt64(int64(currentBit))
	}

	return db
}
