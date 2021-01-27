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

type Bytes struct {
	Entries      [][]byte
	DBLengthSqrt int // unused for vector
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
	var tmpb byte
	entries := make([][]field.Element, numRows)
	numColumns := dbLen / numRows
	for i := 0; i < numRows; i++ {
		entries[i] = make([]field.Element, numColumns)
		for j := 0; j < numColumns; j++ {
			tmp.SetRandom(rnd)
			tmpb = tmp.Bytes()[len(tmp.Bytes())-1]
			if tmpb>>7 == 1 {
				entries[i][j].SetOne()
			} else {
				entries[i][j].SetZero()
			}
		}
	}
	return &DB{Entries: entries, Info: Info{NumColumns: numColumns, NumRows: numRows, BlockSize: 0}}
}

/*
func CreateRandomMultiBitMatrix(rnd io.Reader, dbLen, blockLen int) *DB {
	var err error
	// compute square root of db length
	dbLengthSqrt := math.Sqrt(float64(dbLen))
	if dbLengthSqrt != math.Floor(dbLengthSqrt) {
		log.Fatal(errors.New("square root of db length is not an integer"))
	}
	dbLengthSqrtInt := int(dbLengthSqrt)

	numFieldElements := dbLengthSqrtInt / (field.Bytes * blockLen)
	entries := make([][][]field.Element, numFieldElements)
	for i := 0; i < dbLengthSqrtInt; i++ {
		entries[i] = make([][]field.Element, numFieldElements)
		for j := 0; j < dbLengthSqrtInt; j++ {
			entries[i][j], err = field.RandomVector(rnd, blockLen)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return &DB{Entries: entries, NumColumns: dbLengthSqrtInt}
}

func CreateMultiBitGFLength(length int) *DB {
	entries := make([][][]field.Element, constants.DBLength)
	for i := range entries {
		entries[i] = make([][]field.Element, 1)
		for j := range entries[i] {
			entries[i][j] = zeroVectorGF(length)
		}
	}

	return &DB{Entries: entries}
}

func CreateMultiBitGF() *DB {
	entries := make([][]field.Element, constants.DBLength)
	one := field.One()
	toRetrieve := []field.Element{one, one, one, one, one, one, one, one, one, one, one, one, one, one, one, one}
	for i := range entries {
		if i == 0 {
			entries[i] = toRetrieve
		} else {
			entries[i] = zeroVectorGF(cst.BlockLength)
		}
	}

	return &DB{Entries: entries}
}

func CreateVectorGF() *DB {
	entries := make([][]field.Element, 1)
	entries[0] = zeroVectorGF(cst.DBLength)

	return &DB{Entries: entries}
}

func CreateAsciiVectorGF() *DB {
	// playing with VPIR in ascii
	db := CreateVectorGF()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		if currentBit == 0 {
			val := field.Zero()
			db.Entries[0][i] = val
		} else {
			val := field.One()
			db.Entries[0][i] = val
		}

	}

	return db
}

func CreateMatrixGF() *DB {
	// compute square root of db length
	dbLengthSqrt := math.Sqrt(cst.DBLength)
	if dbLengthSqrt != math.Floor(dbLengthSqrt) {
		panic(errors.New("Square root of db length is not an integer"))
	}
	dbLengthSqrtInt := int(dbLengthSqrt)

	entries := make([][]field.Element, dbLengthSqrtInt)
	for i := 0; i < dbLengthSqrtInt; i++ {
		entries[i] = zeroVectorGF(dbLengthSqrtInt)
	}

	return &DB{Entries: entries, NumColumns: dbLengthSqrtInt}
}

func CreateAsciiMatrixGF() *DB {
	// playing with VPIR in ascii
	db := CreateMatrixGF()
	for i, b := range text {
		currentBit, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		entry := field.Zero()
		if currentBit == 1 {
			entry = field.One()
		}
		db.Entries[i/db.NumColumns][i%db.NumColumns] = entry
	}

	return db
}

func CreateAsciiMatrixOneKb() *DB {
	data := make([]byte, 1024)
	rand.Read(data)
	db := CreateMatrixGF()

	bits := utils.Bytes2Bits(data)

	for i, b := range bits {
		entry := field.Zero()
		if b == 1 {
			entry = field.One()
		}
		db.Entries[i/db.NumColumns][i%db.NumColumns] = entry
	}

	return db
}*/

func zeroVectorGF(length int) []field.Element {
	v := make([]field.Element, length)
	for i := 0; i < length; i++ {
		v[i] = field.Zero()
	}

	return v
}

/*
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

	return &Bytes{Entries: entries, NumColumns: dbLengthSqrtInt}
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
		db.Entries[i/db.NumColumns][i%db.NumColumns] = entry
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
		db.Entries[i/db.NumColumns][i%db.NumColumns] = entry
	}

	return db
}*/
