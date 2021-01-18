package database

import (
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

func GenerateRandomDB(path string) (*DB, int, int, int, error) {
	// parse id->key file
	pairs, err := utils.ParseCSVRandomIDKeys(path)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// analyze keys
	idLength, keyLength := utils.AnalyzeIDKeys(pairs)
	idLength = constants.IDLengthBytes
	entryLength := idLength + keyLength

	// TODO: numRows should be picked algorithmically
	numRows := constants.DBLength

	// generate hash table
	hashTable, err := generateHashTable(pairs, numRows, idLength)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// get maximal []byte length in hashTable
	maxBytes := utils.MaxBytesLength(hashTable)

	chunkLength := constants.ChunkBytesLength

	// compute field elements necessary to encode the maximum length
	blockLength := int(math.Ceil(float64(maxBytes) / float64(chunkLength)))

	// create all zeros db
	db := CreateMultiBitGFLength(blockLength)

	// embed data into field elements
	for id, v := range hashTable {
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

		// store in db and automatically pad
		copy(db.Entries[id], elements)
	}

	return db, idLength, keyLength, blockLength, nil
}

func generateHashTable(pairs map[string][]byte, numRows, idLength int) (map[int][]byte, error) {

	// prepare db
	db := make(map[int][]byte)

	// range over all id,key pairs and assign every pair to a given bucket
	for id, k := range pairs {
		hashKey := utils.HashToIndex(id, numRows)

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
