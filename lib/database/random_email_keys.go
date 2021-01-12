package database

import (
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
)

func GenerateRandomDB(path string) (*GF, int) {
	n := 10000
	hashTable := generateHashTable(path)

	// get maximal []byte length in hashTable
	max := 0
	for _, v := range hashTable {
		if len(v) > max {
			max = len(v)
		}
	}

	fieldElementsMax := int(math.Ceil(float64(max) / 15.0))

	// create all zeros db
	// TODO: this is actually useless, but just for testing
	db := CreateMultiBitGFLength(fieldElementsMax)

	// embed data into field elements
	chunkLength := 15
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
		for len(elements) < max {
			elements = append(elements, field.Zero())
		}

		// store in db
		db.Entries[id] = elements
	}

	return db, fieldElementsMax
}

func generateHashTable(path string) (map[int][]byte, error) {
	// parse id->key file
	pairs, err := utils.ParseRandomIDKeys(path)
	if err != nil {
		return nil, err
	}

	// analyze pairs
	maxIDLength, maxKeyLength := utils.AnalyzeIDKeys(pairs)

	// prepare db
	db := make(map[int][]byte)

	// range over all id,key pairs and assign every pair to a given bucket
	for id, k := range pairs {
		hashKey := utils.HashToIndex(id, constants.DBLength)

		// prepare entry
		idBytes := make([]byte, maxIDLength)
		copy(idBytes, id)
		entry = append(idKey, k...)

		if _, ok := db[hashKey]; !ok {
			db[hashKey] = entry
		} else {
			db[hashKey] = append(db[hashKey], entry...)
		}
	}

	return db
}
