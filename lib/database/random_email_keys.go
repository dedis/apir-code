package database

import (
	"hash/maphash"
	"math"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
)

func GenerateRandomDB() (*GF, int) {
	n := 10000
	hashTable := generateHashTable(n)

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

func generateHashTable(n int) map[int][]byte {
	var h maphash.Hash
	pairs := generateRandomIdKeyPairs(n)

	db := make(map[int][]byte)

	for id, k := range pairs {
		hashKey := hashFunction(h, uint64(constants.DBLength), id)
		id += ","
		idKey := make([]byte, 44)
		copy(idKey, id)
		idKey = append(idKey, k...)
		if _, ok := db[hashKey]; !ok {
			db[hashKey] = idKey
		} else {
			db[hashKey] = append(db[hashKey], idKey...)
		}
	}

	return db
}

func hashFunction(h maphash.Hash, dbLength uint64, id string) int {
	h.WriteString(id)
	sum := h.Sum64()
	h.Reset()

	return int(sum % dbLength)
}
