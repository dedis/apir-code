package database

import (
	"errors"
	"math"
	"math/big"
	"strconv"

	cst "github.com/si-co/vpir-code/lib/constants"
)

type Vector struct {
	Entries []*big.Int
}

func CreateVector() *Vector {
	entries := make([]*big.Int, cst.DBLength)
	for i := 0; i < cst.DBLength; i++ {
		entries[i] = cst.BigZero
	}

	return &Vector{Entries: entries}
}

func CreateAsciiVector() *Vector {
	// playing with VPIR in ascii
	text := "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"
	db := CreateVector()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		db.Entries[i] = new(big.Int).SetInt64(int64(currentBit))
	}

	return db
}

type Matrix struct {
	Entries      [][]*big.Int
	DBLengthSqrt int
}

func CreateMatrix() *Matrix {
	// compute square root of db length
	dbLengthSqrt := math.Sqrt(cst.DBLength)
	if dbLengthSqrt != math.Floor(dbLengthSqrt) {
		panic(errors.New("Square root of db length is not an integer"))
	}
	dbLengthSqrtInt := int(dbLengthSqrt)

	entries := make([][]*big.Int, dbLengthSqrtInt)
	for i := 0; i < dbLengthSqrtInt; i++ {
		entries[i] = make([]*big.Int, dbLengthSqrtInt)
		for j := 0; j < dbLengthSqrtInt; j++ {
			entries[i][j] = cst.BigOne
		}
	}
	//entries[9] = cst.BigZero

	return &Matrix{Entries: entries, DBLengthSqrt: dbLengthSqrtInt}
}

func CreateAsciiMatrix() *Matrix {
	// playing with VPIR in ascii
	text := "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"
	db := CreateMatrix()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		entry := new(big.Int).SetInt64(int64(currentBit))
		db.Entries[i/db.DBLengthSqrt][i%db.DBLengthSqrt] = entry
	}

	return db
}
