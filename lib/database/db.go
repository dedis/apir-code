package database

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"strconv"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
)

type VectorGF struct {
	Entries []*field.Element
}

func CreateVectorGF() *VectorGF {
	entries := make([]*field.Element, cst.DBLength)
	for i := 0; i < cst.DBLength; i++ {
		entries[i] = field.Zero()
		// precompute multiplication table
		entries[i].PrecomputeMul()
	}

	return &VectorGF{Entries: entries}
}

func CreateAsciiVectorGF() *VectorGF {
	// playing with VPIR in ascii
	text := "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"
	db := CreateVectorGF()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		if currentBit == 0 {
			db.Entries[i] = field.Zero()
			db.Entries[i].PrecomputeMul()
		} else {
			db.Entries[i] = field.One()
			db.Entries[i].PrecomputeMul()
		}

	}

	return db
}

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

type MatrixGF struct {
	Entries      [][]*field.Element
	DBLengthSqrt int
}

func CreateMatrixGF() *MatrixGF {
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

	return &MatrixGF{Entries: entries, DBLengthSqrt: dbLengthSqrtInt}
}

func CreateAsciiMatrixGF() *MatrixGF {
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

func CreateAsciiMatrixOneKb() *MatrixGF {
	data := make([]byte, 1024)
	rand.Read(data)
	db := CreateMatrixGF()

	bits := Bytes2Bits(data)

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

func Bytes2Bits(data []byte) []int {
	dst := make([]int, 0)
	for _, v := range data {
		for i := 0; i < 8; i++ {
			move := uint(7 - i)
			dst = append(dst, int((v>>move)&1))
		}
	}
	return dst
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
