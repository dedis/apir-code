package database

import (
	"crypto"
	"encoding/binary"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/cloudflare/circl/group"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
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
	BlockLength  int // length of data in blocks defined in number of elements
}

type Info struct {
	NumRows    int
	NumColumns int
	BlockSize  int
	// TODO: remove this, should use the block length defined in the KeyInfo struct
	BlockLengths []int // length of data in blocks defined in number of elements

	// PIR type: classical, merkle, signature
	PIRType string

	*Auth
	*Merkle

	// Lattice parameters for the single-server data retrieval
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
		Info:     info,
		KeysInfo: make([]*KeyInfo, 0),
		Entries:  make([]uint32, 0),
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
	rand.Seed(time.Now().UnixNano())
	entryLength := 2

	// create random keys
	// for random db use 2048 bits = 64 uint32 elements
	entries := field.RandVectorWithPRG(numIdentifiers*entryLength, rnd)

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
		id := packet.NewUserId("", "", utils.Ranstring(16))

		// in this case lengths are all equal, 2048 bits = 64 uint32 elements
		bl := entryLength

		keysInfo[i] = &KeyInfo{
			UserId:       id,
			CreationTime: ct,
			PubKeyAlgo:   pka,
			BlockLength:  bl,
		}
	}

	// in this case lengths are all equal
	// TODO: here for compatibility reasons, FIX
	info := NewInfo(1, numIdentifiers, entryLength)
	for i := range info.BlockLengths {
		info.BlockLengths[i] = entryLength
	}

	return &DB{
		KeysInfo: keysInfo,
		Entries:  entries,
		Info:     info,
	}, nil
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
