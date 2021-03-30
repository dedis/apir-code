package database

import (
	"bytes"
	"math"
	"sort"

	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
)

const numKeysToDBLengthRatio float32 = 0.2

var TotalPadding = 0

func GenerateRealKeyDB(dataPaths []string, elementLength int, rebalanced bool) (*DB, error) {
	keys, err := pgp.LoadKeysFromDisk(dataPaths)
	if err != nil {
		return nil, err
	}

	// Sort the keys by id, higher first, to make sure that
	// all the servers end up with an identical hash table.
	sortById(keys)

	var numColumns, numRows int
	// decide on the length of the hash table
	numBlocks := int(float32(len(keys)) * numKeysToDBLengthRatio)
	if rebalanced {
		utils.IncreaseToNextSquare(&numBlocks)
	}
	ht, err := makeHashTable(keys, numBlocks)

	// get the maximum byte length of the values in the hashTable
	// +1 takes into account the padding 0x80 that is always added.
	maxBytes := utils.MaxBytesLength(ht) + 1
	blockLen := int(math.Ceil(float64(maxBytes) / float64(elementLength)))
	if rebalanced {
		numColumns = int(math.Sqrt(float64(numBlocks)))
		numRows = numColumns
	} else {
		numColumns = numBlocks
		numRows = 1
	}

	// create all zeros db
	db, err := CreateZeroMultiBitDB(numRows, numColumns, blockLen)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	// embed data into field elements
	for _, v := range ht {
		// Pad the block to be a multiple of elementLength
		v = PadBlock(v, elementLength)

		// embed all the bytes
		for j := 0; j < len(v); j += elementLength {
			e := new(field.Element).SetBytes(v[j : j+elementLength])
			db.SetEntry(j/elementLength, *e)
		}
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
	TotalPadding += len(zeros)
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

//func getNTopValuesFromMap(m map[string]int, n int) {
//	// Turning the map into this structure
//	type kv struct {
//		Key   string
//		Value int
//	}
//
//	var ss []kv
//	for k, v := range m {
//		ss = append(ss, kv{k, v})
//	}
//
//	// Then sorting the slice by value, higher first.
//	sort.Slice(ss, func(i, j int) bool {
//		return ss[i].Value > ss[j].Value
//	})
//
//	// Print the x top values
//	for _, kv := range ss[:n] {
//		fmt.Printf("%s, %d\n", kv.Key, kv.Value)
//	}
//}
