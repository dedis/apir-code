package database

import (
	"encoding/binary"
	"log"
	"math"
  "sort"

	"golang.org/x/crypto/blake2b"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

func GenerateKeyDB(path string, chunkLength, numRows, numColumns int) (*DB, error) {
	numBlocks := numRows * numColumns

	// parse k->key file
	pairs, err := utils.ParseCSVRandomIDKeys(path)
	if err != nil {
		return nil, err
	}

	// analyze keys
	idLength, keyLength := utils.AnalyzeIDKeys(pairs)
	idLength = constants.IDLengthBytes
	entryLength := idLength + keyLength

	// generate hash table
	hashTable, err := generateHashTable(pairs, numBlocks, idLength)
	if err != nil {
		return nil, err
	}

	// get maximal []byte length in hashTable
	maxBytes := utils.MaxBytesLength(hashTable)

	// Define blockLen as the size of the biggest hash table value;
	// all the other HT values will be padded to the blockLen size
	maxEntries := maxBytes / entryLength
	blockLen := int(math.Ceil(float64(entryLength)/float64(chunkLength))) * maxEntries

	log.Printf("numRows: %d, numColumns: %d, blockLen: %d\n", numRows, numColumns, blockLen)
	// create all zeros db
	db := CreateZeroMultiBitDB(numRows, numColumns, blockLen)

	// add embedding informations to db
	db.IDLength = idLength
	db.KeyLength = keyLength

	// embed data into field elements
	for k, v := range hashTable {
		elements := make([]field.Element, 0)
		// loop over all entries in v to avoid mixing bytes in element
		for i := 0; i < len(v); i += entryLength {
			entry := v[i : i+entryLength]
			// embed all bytes
			for j := 0; j < len(entry); j += chunkLength {
				end := j + chunkLength
				if end > len(entry) {
					end = len(entry)
				}
				e := new(field.Element).SetBytes(entry[j:end])
        e.SetBytes(entry[j:j+1])
				elements = append(elements, *e)
			}
		}
		// store in db last block and automatically pad since we start
		// with an all zeros db
		//fmt.Println("len in db:", len(db.Entries[k/numColumns][k%numColumns]), "blockLen:", blockLen, "len(elements):", len(elements), "len v:", len(v))
		copy(db.Entries[k/numColumns][k%numColumns], elements)
	}

	return db, nil
}

func generateHashTable(pairs map[string][]byte, maxNumHashKeys, idLength int) (map[int][]byte, error) {

	// prepare db
	db := make(map[int][]byte)

  keys := make([]string, 0)
  for k, _ := range pairs {
    keys = append(keys, k)
  }

  sort.Strings(keys)

	// range over all id,v pairs and assign every pair to a given bucket
	for _, id := range keys {
    v := pairs[id]
		hashKey := HashToIndex(id, maxNumHashKeys)

		// prepare entry
		idBytes := make([]byte, idLength)
		copy(idBytes, id)
		entry := append(idBytes, v...)

		if _, ok := db[hashKey]; !ok {
			db[hashKey] = entry
		} else {
			db[hashKey] = append(db[hashKey], entry...)
		}
	}

	return db, nil
}

// HashToIndex hashes the given id to an index for a database of the given
// length
func HashToIndex(id string, length int) int {
	hash := blake2b.Sum256([]byte(id))
	return int(binary.BigEndian.Uint64(hash[:]) % uint64(length))
}
