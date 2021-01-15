package database

import (
	"fmt"
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

func GenerateRandomDB(path string) (*GF, int, int, error) {
	// parse id->key file
	pairs, err := utils.ParseCSVRandomIDKeys(path)
	if err != nil {
		return nil, 0, 0, err
	}

	// analyze pairs
	maxIDLength, maxKeyLength := utils.AnalyzeIDKeys(pairs)
	fmt.Println(maxIDLength)
	maxIDLength = 45
	entryLength := maxIDLength + maxKeyLength

	fmt.Println(entryLength)

	// generate hash table
	hashTable, err := generateHashTable(pairs, maxIDLength)
	if err != nil {
		return nil, 0, 0, err
	}

	// get maximal []byte length in hashTable
	maxBytes := 0
	for _, v := range hashTable {
		if len(v) > maxBytes {
			maxBytes = len(v)
		}
	}

	// use maximal chunk length
	chunkLength := 15

	// compute field elements necessary to encode the maximum length
	fieldElementsMax := int(math.Ceil(float64(maxBytes) / float64(chunkLength)))
	entryLengthPadding := int(math.Ceil(float64(entryLength)/float64(chunkLength))) * 15

	// compute number of field elements to encode an entry and the bytes of data in last
	// field element
	//fieldElementsEntry := int(math.Ceil(float64(entryLength) / float64(chunkLength)))
	bytesLastFieldElement := entryLength % chunkLength
	fmt.Println(bytesLastFieldElement)

	// create all zeros db
	// TODO: useless to create an all zero db, better to build the db from scratch
	db := CreateMultiBitGFLength(fieldElementsMax)

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

		// pad to have a full block
		for len(elements) < fieldElementsMax {
			elements = append(elements, field.Zero())
		}

		// store in db
		db.Entries[id] = elements
	}

	return db, entryLengthPadding, fieldElementsMax, nil
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
