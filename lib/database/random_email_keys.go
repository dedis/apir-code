package database

/*import (
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

func GenerateRandomDB(path string) (*DB, int, int, error) {
	// parse id->key file
	pairs, err := utils.ParseCSVRandomIDKeys(path)
	if err != nil {
		return nil, 0, 0, err
	}

	// analyze pairs
	maxIDLength, maxKeyLength := utils.AnalyzeIDKeys(pairs)
	entryLength := maxIDLength + maxKeyLength

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

	// compute number of field elements to encode an entry and the bytes of data in last
	// field element
	fieldElementsEntry := int(math.Ceil(float64(entryLength) / float64(chunkLength)))
	bytesLastFieldElement := entryLength % chunkLength

	// create all zeros db
	// TODO: useless to create an all zero db, better to build the db from scratch
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

	return db, fieldElementsEntry, bytesLastFieldElement, nil
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
}*/
