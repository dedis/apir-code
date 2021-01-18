package database

import (
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

func GenerateRandomDB(path string) (*DB, int, int, error) {
	// maximum numer of bytes embedded in a field elements
	chunkLength := constants.ChunkBytesLength

	// parse id->key file
	pairs, err := utils.ParseCSVRandomIDKeys(path)
	if err != nil {
		return nil, 0, 0, err
	}

	// analyze keys
	idLength, keyLength := utils.AnalyzeIDKeys(pairs)
	idLength = constants.IDLengthBytes
	entryLength := idLength + keyLength

	// TODO: these should be picked algorithmically
	numRows := constants.DBLength
	blockLength := constants.BlockLength

	// generate hash table
	hashTable, err := generateHashTable(pairs, numRows, idLength)
	if err != nil {
		return nil, 0, 0, err
	}

	// get maximal []byte length in hashTable
	maxBytes := utils.MaxBytesLength(hashTable)

	// compute field elements necessary to encode the maximum length
	numColumns := int(math.Ceil(float64(maxBytes) / float64(chunkLength) *
		float64(blockLength)))

	// create all zeros db
	db := CreateZeroMultiBitDB(numRows, numColumns, blockLength)

	// embed data into field elements
	for id, v := range hashTable {
		elements := make([]field.Element, 0)
		// count filled blocks
		j := 0
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
				if len(elements) == blockLength {
					copy(db.Entries[id][j], elements)
					elements = make([]field.Element, 0)
					j++
				}
			}
		}

		// store in db last block and automatically pad since we start
		// with an all zeros db
		copy(db.Entries[id][j], elements)
	}

	return db, idLength, keyLength, nil
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
