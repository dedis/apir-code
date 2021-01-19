package database

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"math"

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
	blockLen := int(math.Ceil(float64(maxBytes) / float64(chunkLength)))

	fmt.Println("numRows:", numRows, "numColumns:", numColumns, "blockLen:", blockLen)
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
			for i := 0; i < len(entry); i += chunkLength {
				end := i + chunkLength
				if end > len(entry) {
					end = len(entry)
				}
				e := new(field.Element).SetBytes(entry[i:end])
				elements = append(elements, *e)
			}
		}
		// store in db last block and automatically pad since we start
		// with an all zeros db
		copy(db.Entries[k / numColumns][k % numColumns], elements)
	}

	return db, nil
}

func generateHashTable(pairs map[string][]byte, numMapKeys, idLength int) (map[int][]byte, error) {

	// prepare db
	db := make(map[int][]byte)

	// range over all id,key pairs and assign every pair to a given bucket
	for id, k := range pairs {
		hashKey := HashToIndex(id, numMapKeys)

		// prepare entry
		idBytes := make([]byte, idLength)
		copy(idBytes, id)
		entry := append(idBytes, k...)

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