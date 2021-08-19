package database

import (
	"bytes"
	"log"
	"sort"

	"golang.org/x/crypto/blake2b"

	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/merkle"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/utils"
)

const (
	hashLength                     = 32
	numKeysToDBLengthRatio float32 = 0.1
)

func GenerateRealKeyDB(dataPaths []string, elementLength int, rebalanced bool) (*DB, error) {
	log.Printf("Field db rebalanced: %v, loading keys: %v\n", rebalanced, dataPaths)

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

	info := NewInfo(numRows, numColumns, maxKeyLength(keys))
	db, err := NewEmptyDB(info)
	if err != nil {
		return nil, err
	}

	db.IdLen = hashLength
	for i := 0; i < len(keys); i++ {
		db.Identifiers = append(db.Identifiers, IdToHash(keys[i].ID)...)
		db.Entries = append(db.Entries, field.ByteSliceToFieldElementSlice(keys[i].Packet)...)
		db.BlockLengths[i] = len(keys[i].Packet)
	}

	//ht := makeHashTable(keys, numRows*numColumns)
	//// get the maximum byte length of the values in the hashTable
	//// +1 takes into account the padding 0x80 that is always added.
	//maxBytes := utils.MaxBytesLength(ht) + 1
	//blockLen := int(math.Ceil(float64(maxBytes) / float64(elementLength)))
	//
	//// embed data into field elements
	//blocks := make([][]field.Element, numRows*numColumns)
	//totalLength := 0
	//for k, v := range ht {
	//	// Pad the block to be a multiple of elementLength
	//	v = PadBlock(v, elementLength)
	//	elements := make([]field.Element, len(v)/elementLength)
	//
	//	// embed all the bytes
	//	for j := 0; j < len(v); j += elementLength {
	//		e := new(field.Element).SetBytes(v[j : j+elementLength])
	//		elements[j/elementLength] = *e
	//	}
	//	blocks[k] = elements
	//	totalLength += len(elements)
	//}
	//
	//// create all zeros db
	//db, err := InitMultiBitDBWithCapacity(numRows, numColumns, blockLen, totalLength)
	//if err != nil {
	//	return nil, xerrors.Errorf("failed to create zero db: %v", err)
	//}
	//
	//for k, block := range blocks {
	//	db.AppendBlock(block)
	//	db.BlockLengths[k] = len(block)
	//}

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
	db := InitMultiBitBytes(numRows, numColumns, blockLen)

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
			// BlockSize here is simply to differentiate from the single-bit scheme,
			// not used otherwise.
			BlockSize: maxBlockLen,
			PIRType:   "merkle",
			Merkle:    &Merkle{Root: tree.Root(), ProofLen: proofLen},
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

func IdToHash(id string) []byte {
	hash := blake2b.Sum256([]byte(id))
	return hash[:]
}
