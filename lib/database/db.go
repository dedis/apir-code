package database

import (
	"crypto/rand"
	"errors"
	"math"
	"strconv"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

type GF struct {
	Entries      [][]*field.Element
	DBLengthSqrt int // unused for vector
}

func CreateVectorGF() *GF {
	entries := make([][]*field.Element, 1)
	entries[0] = make([]*field.Element, cst.DBLength)
	for i := 0; i < cst.DBLength; i++ {
		entries[0][i] = field.Zero()
		// precompute multiplication table
		entries[0][i].PrecomputeMul()
	}

	return &GF{Entries: entries}
}

func CreateAsciiVectorGF() *GF {
	// playing with VPIR in ascii
	text := "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"
	db := CreateVectorGF()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		if currentBit == 0 {
			db.Entries[0][i] = field.Zero()
			db.Entries[0][i].PrecomputeMul()
		} else {
			db.Entries[0][i] = field.One()
			db.Entries[0][i].PrecomputeMul()
		}

	}

	return db
}

func CreateMatrixGF() *GF {
	// compute square root of db length
	dbLengthSqrt := math.Sqrt(cst.DBLength)
	if dbLengthSqrt != math.Floor(dbLengthSqrt) {
		panic(errors.New("Square root of db length is not an integer"))
	}
	dbLengthSqrtInt := int(dbLengthSqrt)

	entries := make([][]*field.Element, dbLengthSqrtInt)
	for i := 0; i < dbLengthSqrtInt; i++ {
		entries[i] = make([]*field.Element, dbLengthSqrtInt)
		for j := 0; j < dbLengthSqrtInt; j++ {
			entries[i][j] = field.Zero()
			entries[i][j].PrecomputeMul()
		}
	}

	return &GF{Entries: entries, DBLengthSqrt: dbLengthSqrtInt}
}

func CreateAsciiMatrixGF() *GF {
	// playing with VPIR in ascii
	text := "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"
	db := CreateMatrixGF()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		entry := field.Zero()
		entry.PrecomputeMul()
		if currentBit == 1 {
			entry = field.One()
			entry.PrecomputeMul()
		}
		db.Entries[i/db.DBLengthSqrt][i%db.DBLengthSqrt] = entry
	}

	return db
}

func CreateAsciiMatrixOneKb() *GF {
	data := make([]byte, 1024)
	rand.Read(data)
	db := CreateMatrixGF()

	bits := utils.Bytes2Bits(data)

	for i, b := range bits {
		entry := field.Zero()
		entry.PrecomputeMul()
		if b == 1 {
			entry = field.One()
			entry.PrecomputeMul()
		}
		db.Entries[i/db.DBLengthSqrt][i%db.DBLengthSqrt] = entry
	}

	return db
}
