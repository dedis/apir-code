package database

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/utils"
)

const numKeysToDBLengthRatio float32 = 0.2


func GenerateRealKeyDB(dataPath string, numRows, elementLength int) (*DB, error) {
	var keys []*pgp.Key
	var err error
	keys, err = pgp.LoadKeysFromDisk(dataPath)
	// Sort the keys by id, higher first, to make sure that
	// all the servers end up with an identical hash table.
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Id > keys[j].Id
	})
	if err != nil {
		return nil, err
	}
	// decide on the length of the hash table
	tableLen := int(float32(len(keys)) * numKeysToDBLengthRatio)
	ht, err := makeHashTable(keys, tableLen)

	// get the maximum byte length of the values in the hashTable
	// +1 takes into account the padding 0x80 that is always added.
	maxBytes := utils.MaxBytesLength(ht) + 1
	fmt.Println(maxBytes)
	blockLen := int(math.Ceil(float64(maxBytes)/float64(elementLength)))
	numColumns := tableLen

	// create all zeros db
	db := CreateZeroMultiBitDB(numRows, numColumns, blockLen)

	// embed data into field elements
	for k, v := range ht {
		elements := field.ZeroVector(blockLen)
		// Pad the block
		v = PadBlock(v)
		// embed all bytes
		for j := 0; j < len(v); j += elementLength {
			end := j + elementLength
			if end > len(v) {
				end = len(v)
			}
			e := new(field.Element).SetBytes(v[j:end])
			elements[j/elementLength] = *e
		}
		// store in db last block and automatically pad since we start
		// with an all zeros db
		copy(db.Entries[k/numColumns][(k%numColumns)*blockLen:(k%numColumns+1)*blockLen], elements)
	}

	return db, nil
}

func makeHashTable(keys []*pgp.Key, tableLen int) (map[int][]byte, error) {
	// prepare db
	db := make(map[int][]byte)

	// range over all id,v pairs and assign every pair to a given bucket
	for _, key := range keys {
		hashKey := HashToIndex(key.Id, tableLen)
		db[hashKey] = append(db[hashKey], key.Packet...)
	}

	return db, nil
}

// Simple ISO/IEC 7816-4 padding where 0x80 is appended to the block and
// zeros are assumed afterwards
func PadBlock(block []byte) []byte{
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
