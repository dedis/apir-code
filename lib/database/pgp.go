package database

import (
	"bytes"
	"errors"
	"log"
	"sort"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/si-co/vpir-code/lib/merkle"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/utils"
)

const numKeysToDBLengthRatio float32 = 0.1

func GenerateRealKeyDB(dataPaths []string) (*DB, error) {
	log.Printf("Loading keys: %v\n", dataPaths)

	keys, err := pgp.LoadKeysFromDisk(dataPaths)
	if err != nil {
		return nil, err
	}

	// Sort the keys by id, higher first, to make sure that
	// all the servers end up with an identical hash table.
	sortById(keys)

	// only information needed for FSS-based schemes
	info := Info{NumColumns: len(keys)}
	// create empty database
	db := NewKeysDB(info)

	// iterate and embed keys
	for i := 0; i < len(keys); i++ {
		// useless to store keys for FSS queries
		//key := field.BytesToElements(keys[i].Packet)
		//db.Entries = append(db.Entries, key...)

		keyInfo, err := GetKeyInfoFromPacket(keys[i].Packet)
		if err != nil {
			log.Fatalf("error getting info from a key block: %v", err)
		}

		db.KeysInfo = append(db.KeysInfo, keyInfo)
	}

	return db, nil
}

func GenerateRealKeyBytes(dataPaths []string, rebalanced bool) (*Bytes, error) {
	log.Printf("Bytes db rebalanced: %v, loading keys: %v\n", rebalanced, dataPaths)

	keys, err := pgp.LoadKeysFromDisk(dataPaths)
	if err != nil {
		return nil, err
	}
	// Sort the keys by id, higher first, to make sure that
	// all the servers end up with an identical hash table.
	sortById(keys)

	// decide on the length of the hash table
	preSquareNumBlocks := int(float32(len(keys)) * numKeysToDBLengthRatio)
	numRows, numColumns := CalculateNumRowsAndColumns(preSquareNumBlocks, rebalanced)

	ht := makeHashTable(keys, numRows*numColumns)
	// get the maximum byte length of the values in the hashTable
	// +1 takes into account the padding 0x80 that is always added.
	blockLen := utils.MaxBytesLength(ht) + 1

	// create all zeros db
	db := InitBytes(numRows, numColumns, blockLen)

	// order blocks because of map
	blocks := make([][]byte, numRows*numColumns)
	for k, v := range ht {
		// appending only 0x80 (without zeros)
		blocks[k] = PadWithSignalByte(v)
	}

	// add blocks to the db with the according padding and store the length
	for k, block := range blocks {
		db.BlockLengths[k] = len(block)
		db.Entries = append(db.Entries, block...)
	}

	return db, nil
}

func GenerateRealKeyMerkle(dataPaths []string, rebalanced bool) (*Bytes, error) {
	log.Printf("Bytes db rebalanced: %v, loading keys: %v\n", rebalanced, dataPaths)

	keys, err := pgp.LoadKeysFromDisk(dataPaths)
	if err != nil {
		return nil, err
	}
	// Sort the keys by id, higher first, to make sure that
	// all the servers end up with an identical hash table.
	sortById(keys)

	// decide on the length of the hash table
	preSquareNumBlocks := int(float32(len(keys)) * numKeysToDBLengthRatio)
	numRows, numColumns := CalculateNumRowsAndColumns(preSquareNumBlocks, rebalanced)
	ht := makeHashTable(keys, numRows*numColumns)

	// map into blocks
	blocks := make([][]byte, numRows*numColumns)
	for k, v := range ht {
		// appending only 0x80 (without zeros)
		blocks[k] = PadWithSignalByte(v)
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		return nil, err
	}

	proofLen := tree.EncodedProofLength()
	maxBlockLen := 0
	blockLens := make([]int, numRows*numColumns)
	for i := 0; i < numRows*numColumns; i++ {
		// we add +1 for appending 0x80 to the proof
		blockLens[i] = len(blocks[i]) + proofLen + 1
		if blockLens[i] > maxBlockLen {
			maxBlockLen = blockLens[i]
		}
	}

	entries := makeMerkleEntries(blocks, tree, numRows, numColumns, maxBlockLen)

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:      numRows,
			NumColumns:   numColumns,
			BlockLengths: blockLens,
			PIRType:      "merkle",
			Merkle:       &Merkle{Root: tree.Root(), ProofLen: proofLen},
		},
	}

	return m, nil
}

func makeHashTable(keys []*pgp.Key, tableLen int) map[int][]byte {
	// prepare db
	db := make(map[int][]byte)

	// range over all id,v pairs and assign every pair to a given bucket
	for _, key := range keys {
		hashKey := HashToIndex(key.ID, tableLen)
		db[hashKey] = append(db[hashKey], key.Packet...)
	}

	return db
}

// Simple ISO/IEC 7816-4 padding where 0x80 is appended to the block, then
// zeros to make up to blockLen
func PadBlock(block []byte, blockLen int) []byte {
	block = append(block, byte(0x80))
	zeros := make([]byte, blockLen-(len(block)%blockLen))
	return append(block, zeros...)
}

func PadWithSignalByte(block []byte) []byte {
	return append(block, byte(0x80))
}

func UnPadBlock(block []byte) []byte {
	// remove zeros
	block = bytes.TrimRightFunc(block, func(b rune) bool {
		return b == 0
	})
	// remove 0x80 preceding zeros
	return block[:len(block)-1]
}

func sortById(keys []*pgp.Key) {
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].ID > keys[j].ID
	})
}

func maxKeyLength(keys []*pgp.Key) int {
	max := 0
	for i := 0; i < len(keys); i++ {
		if len(keys[i].Packet) > max {
			max = len(keys[i].Packet)
		}
	}

	return max
}

// GetKeyInfoFromPacket parses packet bytes and returns information about the key
func GetKeyInfoFromPacket(pkt []byte) (*KeyInfo, error) {
	// parse the input bytes as a key ring
	reader := bytes.NewReader(pkt)
	el, err := openpgp.ReadKeyRing(reader)
	if err != nil {
		return nil, err
	}
	// the key ring is supposed to have only one Entity
	if len(el) != 1 {
		return nil, errors.New("more than one openpgp entity in a key block")
	}

	return &KeyInfo{
		UserId:       el[0].PrimaryIdentity().UserId,
		CreationTime: el[0].PrimaryKey.CreationTime,
		PubKeyAlgo:   el[0].PrimaryKey.PubKeyAlgo,
	}, nil
}
