package database

import (
	"bytes"
	"log"
	"math"
	"sort"

	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
)

const numKeysToDBLengthRatio float32 = 0.1

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

	ht, err := makeHashTable(keys, numRows*numColumns)
	// get the maximum byte length of the values in the hashTable
	// +1 takes into account the padding 0x80 that is always added.
	maxBytes := utils.MaxBytesLength(ht) + 1
	blockLen := int(math.Ceil(float64(maxBytes) / float64(elementLength)))

	// create all zeros db
	db, err := CreateZeroMultiBitDB(numRows, numColumns, blockLen)
	if err != nil {
		return nil, xerrors.Errorf("failed to create zero db: %v", err)
	}

	// embed data into field elements
	for k, v := range ht {
		// Pad the block to be a multiple of elementLength
		v = PadBlock(v, elementLength)
		elements := make([]field.Element, len(v)/elementLength)

		// embed all the bytes
		for j := 0; j < len(v); j += elementLength {
			e := new(field.Element).SetBytes(v[j : j+elementLength])
			elements[j/elementLength] = *e
		}
		for m := k * blockLen; m < (k+1)*blockLen; m++ {
			if m-k*blockLen >= len(elements) {
				break
			}
			db.SetEntry(m, elements[m-k*blockLen])
		}
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

	ht, err := makeHashTable(keys, numRows*numColumns)
	// get the maximum byte length of the values in the hashTable
	// +1 takes into account the padding 0x80 that is always added.
	blockLen := utils.MaxBytesLength(ht) + 1

	// create all zeros db
	db := CreateZeroMultiBitBytes(numRows, numColumns, blockLen)

	// embed data into bytes
	for k, v := range ht {
		v = PadBlock(v, blockLen)
		copy(db.Entries[k*blockLen:(k+1)*blockLen], v)
	}

	return db, nil
}

func makeHashTable(keys []*pgp.Key, tableLen int) (map[int][]byte, error) {
	// prepare db
	db := make(map[int][]byte)

	// range over all id,v pairs and assign every pair to a given bucket
	for _, key := range keys {
		hashKey := HashToIndex(key.ID, tableLen)
		db[hashKey] = append(db[hashKey], key.Packet...)
	}

	return db, nil
}

// Simple ISO/IEC 7816-4 padding where 0x80 is appended to the block, then
// zeros to make up to blockLen
func PadBlock(block []byte, blockLen int) []byte {
	block = append(block, byte(0x80))
	zeros := make([]byte, blockLen-(len(block)%blockLen))
	return append(block, zeros...)
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
