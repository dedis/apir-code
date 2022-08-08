package database

import (
	"crypto"
	"encoding/binary"
	"io"
	"math"
	"math/rand"
	"time"

	"golang.org/x/xerrors"

	"github.com/cloudflare/circl/group"
	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/tuneinsight/lattigo/v3/bfv"
	"golang.org/x/crypto/blake2b"
)

type DB struct {
	KeysInfo []*KeyInfo
	Entries  []uint32

	Info
}

type KeyInfo struct {
	UserId       *packet.UserId
	CreationTime time.Time
	PubKeyAlgo   packet.PublicKeyAlgorithm
	BitLength    uint16
}

type Info struct {
	NumRows      int
	NumColumns   int
	BlockSize    int
	BlockLengths []int // length of data in blocks defined in number of elements

	// PIR type: classical, merkle, signature
	PIRType string

	*Auth
	*Merkle

	// Lattice parameters for the single-server data retrieval
	LatParams bfv.Parameters
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

func NewKeysDB(info Info) *DB {
	return &DB{
		Info:     info,
		KeysInfo: make([]*KeyInfo, 0),
		Entries:  make([]uint32, 0),
	}
}

func NewBitsDB(info Info) *DB {
	return &DB{
		Info:    info,
		Entries: make([]uint32, info.NumRows*info.NumColumns*info.BlockSize),
	}
}

func CreateRandomBitsDB(rnd io.Reader, dbLen, numRows, blockLen int) (*DB, error) {
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

	numBytesToRead := n*field.Bytes + 1
	randBytes := make([]byte, numBytesToRead)
	if _, err := io.ReadFull(rnd, randBytes[:]); err != nil {
		return nil, xerrors.Errorf("failed to read random randBytes: %v", err)
	}

	db := NewBitsDB(info)
	field.BytesToElements(db.Entries, randBytes)

	// add block lengths also in this case for compatibility
	db.BlockLengths = make([]int, numRows*numColumns)
	for i := 0; i < n; i++ {
		db.BlockLengths[i/blockLen] = blockLen
	}

	return db, nil
}

func CreateRandomKeysDB(rnd io.Reader, numIdentifiers int) (*DB, error) {
	// only used for eval, so fine to init the seed for
	// non-crypto PRG with day
	rand.Seed(int64(1000))

	keysInfo := make([]*KeyInfo, numIdentifiers)
	for i := 0; i < numIdentifiers; i++ {
		// random creation date
		ct := utils.Randate()

		// random algorithm, taken from random permutation of
		// https://pkg.go.dev/golang.org/x/crypto/openpgp/packet#PublicKeyAlgorithm
		algorithms := []packet.PublicKeyAlgorithm{1, 16, 17, 18, 19}
		pka := algorithms[rand.Intn(len(algorithms))]

		// random userd id
		// By convention, this takes the form "Full Name (Comment) <email@example.com>"
		// which is split out in the fields below.
		// For testing purposes, only random email and other fields empty strings
		id := packet.NewUserId("", "", utils.Ranstring(32))

		keysInfo[i] = &KeyInfo{
			UserId:       id,
			CreationTime: ct,
			PubKeyAlgo:   pka,
		}
	}

	// only information needed for FSS-based schemes
	info := Info{NumColumns: numIdentifiers}

	return &DB{
		KeysInfo: keysInfo,
		Info:     info,
	}, nil
}

// HashToIndex hashes the given id to an index for a database of the given
// length
func HashToIndex(id string, length int) uint32 {
	hash := blake2b.Sum256([]byte(id))
	return binary.BigEndian.Uint32(hash[:4]) % uint32(length)
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

func (d *DB) SizeGiB() float64 {
	return float64(len(d.Entries)*16) * 9.313e-10
}
