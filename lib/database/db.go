package database

import (
	"crypto"
	"encoding/binary"
	"io"
	"log"
	"math"

	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

type DB struct {
	Entries []field.Element
	Info
}

type Info struct {
	NumRows    int
	NumColumns int
	BlockSize  int
	// PIR type: classical, merkle, signature
	PIRType string

	*Auth
	*Merkle
	*DataEmbedding
	// Lattice parameters for the single-server setting
	//Lattice *bfv.Parameters
}

// Authentication information for the single-server setting
type Auth struct {
	// The global digest that is a hash of all the row digests. Public.
	Digest []byte
	// ECC group and hash algorithm used for digest computation and PIR itself
	Group group.Group
	Hash  crypto.Hash
	// Due to lack of the size functions in the lib API, we store it in the db info
	ElementSize int
	ScalarSize int
}

// Data embedding info
type DataEmbedding struct {
	IDLength  int
	KeyLength int
}

// The info needed for the Merkle-tree based approach
type Merkle struct {
	Root     []byte
	ProofLen int
}

func CreateZeroMultiBitDB(numRows, numColumns, blockSize int) *DB {
	entries := field.ZeroVector(numRows * numColumns * blockSize)
	return &DB{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockSize,
		},
	}
}

func CreateRandomMultiBitDB(rnd io.Reader, dbLen, numRows, blockLen int) *DB {
	numColumns := dbLen / (8 * field.Bytes * numRows * blockLen)
	// handle very small db
	if numColumns == 0 {
		numColumns = 1
	}
	entries, err := field.RandomVector(rnd, numRows*numColumns*blockLen)
	if err != nil {
		log.Fatal(err)
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
	entries := make([]field.Element, dbLen)
	numColumns := dbLen / numRows
	for i := 0; i < dbLen; i++ {
		tmp.SetRandom(rnd)
		tmpb := tmp.Bytes()[len(tmp.Bytes())-1]
		if tmpb>>7 == 1 {
			entries[i].SetOne()
		} else {
			entries[i].SetZero()
		}
	}
	return &DB{Entries: entries, Info: Info{NumColumns: numColumns, NumRows: numRows, BlockSize: 0}}
}

// HashToIndex hashes the given id to an index for a database of the given
// length
func HashToIndex(id string, length int) int {
	hash := blake2b.Sum256([]byte(id))
	return int(binary.BigEndian.Uint64(hash[:]) % uint64(length))
}

func CalculateNumRowsAndColumns(numBlocks int, matrix bool) (numRows, numColumns int) {
	utils.IncreaseToNextSquare(&numBlocks)
	if matrix {
		numColumns = int(math.Sqrt(float64(numBlocks)))
		numRows = numColumns
	} else {
		numColumns = numBlocks
		numRows = 1
	}
	return
}
