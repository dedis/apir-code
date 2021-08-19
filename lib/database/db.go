package database

import (
	"crypto"
	"encoding/binary"
	"io"
	"math"

	"github.com/cloudflare/circl/group"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

type DB struct {
	Identifiers []byte
	Entries     []uint32

	Info
}

type Info struct {
	NumRows    int
	NumColumns int
	BlockSize  int
	// The true length of data in each block,
	// defined in the number of elements
	BlockLengths []int
	IdLen        int // length of each identifier

	// PIR type: classical, merkle, signature
	PIRType string

	*Auth
	*Merkle

	//Lattice parameters for the single-server data retrieval
	LatParams *bfv.Parameters
}

// Auth is authentication information for the single-server setting
type Auth struct {
	// The global digest that is a hash of all the row digests. Public.
	Digest []byte
	// One digest per row, authenticating all the elements in that row.
	SubDigests []byte
	// ECC group and hash algorithm used for digest computation and PIR itself
	Group group.Group
	Hash  crypto.Hash
	// Due to lack of the size functions in the lib API, we store it in the db info
	ElementSize int
	ScalarSize  int
}

// Merkle is the info needed for the Merkle-tree based approach
type Merkle struct {
	Root     []byte
	ProofLen int
}

func NewEmptyDB(info Info) (*DB, error) {
	return &DB{
		Info:        info,
		Identifiers: make([]byte, 0),
		Entries:     make([]uint32, 0),
	}, nil
}

func NewInfo(nRows, nCols, bSize int) Info {
	return Info{
		NumRows:      nRows,
		NumColumns:   nCols,
		BlockSize:    bSize,
		BlockLengths: make([]int, nRows*nCols),
	}
}

func CreateRandomDB(rnd io.Reader, numIdentifiers int) (*DB, error) {
	//identifiers := make([]byte, numIdentifiers*constants.IdentifierLength)
	//if _, err := io.ReadFull(rnd, identifiers[:]); err != nil {
	//return nil, xerrors.Errorf("failed to read random bytes: %v", err)
	//}
	idUint32 := make([]uint32, numIdentifiers)
	for i := range idUint32 {
		idUint32[i] = uint32(i)
	}
	identifiers := utils.Uint32SliceToByteSlice(idUint32)

	// for random db use 2048 bits = 64 uint32 elements
	entryLength := 64
	entries := field.RandVectorWithPRG(numIdentifiers*entryLength, rnd)

	// in this case lengths are all equal
	info := NewInfo(1, numIdentifiers, entryLength)
	for i := range info.BlockLengths {
		info.BlockLengths[i] = entryLength
	}

	return &DB{
		Identifiers: identifiers,
		Entries:     entries,
		Info:        info,
	}, nil
}

//func CreateRandomSingleBitDB(rnd io.Reader, dbLen, numRows int) (*DB, error) {
//	numColumns := dbLen / numRows
//
//	info := Info{
//		NumColumns: numColumns,
//		NumRows:    numRows,
//		BlockSize:  constants.SingleBitBlockLength,
//	}
//
//	db, err := NewDB(info)
//	if err != nil {
//		return nil, xerrors.Errorf("failed to create db: %v", err)
//	}
//
//	buf := make([]byte, dbLen)
//	if _, err := io.ReadFull(rnd, buf[:]); err != nil {
//		return nil, xerrors.Errorf("failed to read random buf: %v", err)
//	}
//
//	for i := 0; i < dbLen; i++ {
//		element := uint32(0)
//		if buf[i]>>7 == 1 {
//			element = 1
//		}
//
//		db.SetEntry(i, element)
//	}
//
//	return db, nil
//}

// HashToIndex hashes the given id to an index for a database of the given
// length
func HashToIndex(id string, length int) int {
	hash := blake2b.Sum256([]byte(id))
	return int(binary.BigEndian.Uint64(hash[:]) % uint64(length))
}

func CalculateNumRowsAndColumns(numBlocks int, matrix bool) (numRows, numColumns int) {
	if matrix {
		utils.IncreaseToNextSquare(&numBlocks)
		numColumns = int(math.Sqrt(float64(numBlocks)))
		numRows = numColumns
	} else {
		numColumns = numBlocks
		numRows = 1
	}
	return
}

/*
func (d *DB) SetEntry(i int, el uint32) {
	d.Entries[i] = []byte(el)
}

func (d *DB) AppendBlock(bl []uint32) {
	d.Entries = append(d.Entries, bl...)
}

func (d *DB) GetEntry(i int) uint32 {
	return d.Entries[i]
}

*/
func (d *DB) Range(begin, end int) []uint32 {
	return d.Entries[begin:end]
}

/*
func InitMultiBitDBWithCapacity(numRows, numColumns, blockSize, cap int) (*DB, error) {
	info := Info{NumColumns: numColumns,
		NumRows:   numRows,
		BlockSize: blockSize,
	}

	db, err := NewEmptyDBWithCapacity(info, cap)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	db.BlockLengths = make([]int, numRows*numColumns)

	return db, nil
}

/*
func InitMultiBitDB(numRows, numColumns, blockSize int) (*DB, error) {
	info := Info{NumColumns: numColumns,
		NumRows:   numRows,
		BlockSize: blockSize,
	}

	db, err := NewEmptyDB(info)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	return db, nil
}

func CreateRandomMultiBitDB(rnd io.Reader, dbLen, numRows, blockLen int) (*DB, error) {
	numColumns := dbLen / (8 * field.Bytes * numRows * blockLen)
	// handle very small db
	if numColumns == 0 {
		numColumns = 1
	}

	info := Info{
		NumColumns: numColumns,
		NumRows:    numRows,
		BlockSize:  blockLen,
	}

	n := numRows * numColumns * blockLen

	bytesLength := n*field.Bytes + 1
	bytes := make([]byte, bytesLength)

	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, xerrors.Errorf("failed to read random bytes: %v", err)
	}

	db, err := NewDB(info)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	// add block lengths also in this case for compatibility
	db.BlockLengths = make([]int, numRows*numColumns)

	for i := 0; i < n; i++ {
		element := binary.BigEndian.Uint32(bytes[i*field.Bytes:(i+1)*field.Bytes]) % field.ModP
		db.SetEntry(i, element)
		db.BlockLengths[i/blockLen] = blockLen
	}

	return db, nil
}
*/
