package database

import (
	"crypto"
	"encoding/binary"
	"io"
	"math"

	"github.com/cloudflare/circl/group"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"
)

func NewDB(info Info) (*DB, error) {
	n := info.BlockSize * info.NumColumns * info.NumRows
	if info.BlockSize == constants.SingleBitBlockLength {
		n = info.NumColumns * info.NumRows
	}

	return &DB{
		Info:     info,
		inMemory: make([]uint32, n),
	}, nil
}

func NewEmptyDBWithCapacity(info Info, cap int) (*DB, error) {
	return &DB{
		Info:     info,
		inMemory: make([]uint32, 0, cap),
	}, nil
}

func NewEmptyDB(info Info) (*DB, error) {
	n := info.BlockSize * info.NumColumns * info.NumRows
	if info.BlockSize == constants.SingleBitBlockLength {
		n = info.NumColumns * info.NumRows
	}

	return NewEmptyDBWithCapacity(info, n)
}

type DB struct {
	Info
	inMemory []uint32
}

func (d *DB) SetEntry(i int, el uint32) {
	d.inMemory[i] = el
}

func (d *DB) AppendBlock(bl []uint32) {
	d.inMemory = append(d.inMemory, bl...)
}

func (d *DB) GetEntry(i int) uint32 {
	return d.inMemory[i]
}

func (d *DB) Range(begin, end int) []uint32 {
	return d.inMemory[begin:end]
}

type Info struct {
	NumRows    int
	NumColumns int
	BlockSize  int
	// The true length of data in each block,
	// defined in the number of elements
	BlockLengths []int

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

func CreateZeroMultiBitDB(numRows, numColumns, blockSize int) (*DB, error) {
	info := Info{NumColumns: numColumns,
		NumRows:   numRows,
		BlockSize: blockSize,
	}

	// already initialized at zero
	db, err := NewDB(info)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	return db, nil
}

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

func CreateRandomSingleBitDB(rnd io.Reader, dbLen, numRows int) (*DB, error) {
	numColumns := dbLen / numRows

	info := Info{
		NumColumns: numColumns,
		NumRows:    numRows,
		BlockSize:  constants.SingleBitBlockLength,
	}

	db, err := NewDB(info)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	buf := make([]byte, dbLen)
	if _, err := io.ReadFull(rnd, buf[:]); err != nil {
		return nil, xerrors.Errorf("failed to read random buf: %v", err)
	}

	for i := 0; i < dbLen; i++ {
		element := uint32(0)
		if buf[i]>>7 == 1 {
			element = 1
		}

		db.SetEntry(i, element)
	}

	return db, nil
}

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
