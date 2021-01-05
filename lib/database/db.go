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

var text = "0101000001101100011000010111100101101001011011100110011100100000011101110110100101110100011010000010000001010110010100000100100101010010"

type GF struct {
	Entries      [][]field.PrecompElement
	DBLengthSqrt int // unused for vector
}

type Bytes struct {
	Entries      [][]byte
	DBLengthSqrt int // unused for vector
}

func CreateVectorGF() *GF {
	entries := make([][]field.PrecompElement, 1)
	entries[0] = zeroVectorGF(cst.DBLength)

	return &GF{Entries: entries}
}

func CreateAsciiVectorGF() *GF {
	// playing with VPIR in ascii
	db := CreateVectorGF()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		if currentBit == 0 {
      val := field.Zero()
			db.Entries[0][i] = val.PrecomputeMul()
		} else {
      val := field.One()
			db.Entries[0][i] = val.PrecomputeMul()
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

	entries := make([][]field.PrecompElement, dbLengthSqrtInt)
	for i := 0; i < dbLengthSqrtInt; i++ {
		entries[i] = zeroVectorGF(dbLengthSqrtInt)
	}

	return &GF{Entries: entries, DBLengthSqrt: dbLengthSqrtInt}
}

func CreateAsciiMatrixGF() *GF {
	// playing with VPIR in ascii
	db := CreateMatrixGF()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		entry := field.Zero().PrecomputeMul()
		if currentBit == 1 {
      entry = field.One().PrecomputeMul()
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
		entry := field.Zero().PrecomputeMul()
		if b == 1 {
			entry = field.One().PrecomputeMul()
		}
		db.Entries[i/db.DBLengthSqrt][i%db.DBLengthSqrt] = entry
	}

	return db
}

func zeroVectorGF(length int) []field.PrecompElement {
	v := make([]field.PrecompElement, length)
	for i := 0; i < length; i++ {
    t := field.Zero()
		v[i] = t.PrecomputeMul()
	}

	return v
}

func CreateVectorByte() *Bytes {
	entries := make([][]byte, 1)
	entries[0] = make([]byte, cst.DBLength)

	return &Bytes{Entries: entries}
}

func CreateAsciiVectorByte() *Bytes {
	// playing with VPIR in ascii
	db := CreateVectorByte()

	zero := byte(0)
	one := byte(1)

	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		if currentBit == 0 {
			db.Entries[0][i] = zero
		} else {
			db.Entries[0][i] = one
		}

	}

	return db
}

func CreateMatrixByte() *Bytes {
	// compute square root of db length
	dbLengthSqrt := math.Sqrt(cst.DBLength)
	if dbLengthSqrt != math.Floor(dbLengthSqrt) {
		panic(errors.New("Square root of db length is not an integer"))
	}
	dbLengthSqrtInt := int(dbLengthSqrt)

	entries := make([][]byte, dbLengthSqrtInt)
	for i := 0; i < dbLengthSqrtInt; i++ {
		entries[i] = make([]byte, dbLengthSqrtInt)
	}

	return &Bytes{Entries: entries, DBLengthSqrt: dbLengthSqrtInt}
}

func CreateAsciiMatrixByte() *Bytes {
	// playing with VPIR in ascii
	db := CreateMatrixByte()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		entry := byte(0)
		if currentBit == 1 {
			entry = byte(1)
		}
		db.Entries[i/db.DBLengthSqrt][i%db.DBLengthSqrt] = entry
	}

	return db
}

func CreateAsciiMatrixOneKbByte() *Bytes {
	data := make([]byte, 1024)
	rand.Read(data)
	db := CreateMatrixByte()

	bits := utils.Bytes2Bits(data)

	for i, b := range bits {
		entry := byte(0)
		if b == 1 {
			entry = byte(1)
		}
		db.Entries[i/db.DBLengthSqrt][i%db.DBLengthSqrt] = entry
	}

	return db
}
