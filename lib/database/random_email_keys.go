package database

import (
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

func GenerateRandomDB(path string) (*GF, error) {
	// parse id->key file
	pairs, err := utils.ParseCSVRandomIDKeys(path)
	if err != nil {
		return nil, err
	}

	// analyze pairs
	maxIDLength, maxKeyLength := utils.AnalyzeIDKeys(pairs)

	// generate hash table
	hashTable, err := generateHashTable(pairs, maxIDLength)
	if err != nil {
		return nil, err
	}

	// get maximal []byte length in hashTable
	maximalEntryLength := maxIDLength + maxKeyLength

	chunkLength := 15
	fieldElementsMax := int(math.Ceil(float64(maximalEntryLength) / float64(chunkLength)))

	// create all zeros db
	// TODO: this is actually useless, but just for testing
	db := CreateMultiBitGFLength(fieldElementsMax)

	// embed data into field elements
	for id, v := range hashTable {
		elements := make([]field.Element, 0)

		// embed all bytes
		for i := 0; i < len(v); i += chunkLength {
			end := i + chunkLength
			if end > len(v) {
				end = len(v)
			}
			e := new(field.Element).SetBytes(v[i:end])
			elements = append(elements, *e)
		}

		// pad to have a full block
		for len(elements) < fieldElementsMax {
			elements = append(elements, field.Zero())
		}

		// store in db
		db.Entries[id] = elements
	}

	return db, nil
}

func generateHashTable(pairs map[string][]byte, maxIDLength int) (map[int][]byte, error) {

	// prepare db
	db := make(map[int][]byte)

	// range over all id,key pairs and assign every pair to a given bucket
	for id, k := range pairs {
		hashKey := utils.HashToIndex(id, constants.DBLength)

		// prepare entry
		idBytes := make([]byte, maxIDLength)
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
