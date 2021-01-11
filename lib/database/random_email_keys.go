package database

import (
	"crypto/rand"
	"hash/maphash"
	"math"

	"github.com/Pallinder/go-randomdata"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
)

func GenerateRandomDB() *GF {
	n := 200
	hashTable := generateHashTable(n)

	// get maximal []byte length in hashTable
	max := 0
	for _, v := range hashTable {
		if len(v) > max {
			max = len(v)
		}
	}

	// create all zeros db
	// TODO: this is actually useless, but just for testing
	db := CreateMultiBitGFLength(int(math.Ceil(float64(max) / 15.0)))

	// embed data into field elements
	chunkLength := 15
	for id, v := range hashTable {
		elements := make([]field.Element, 0)

		// embed all bytes
		for i := 0; i < len(v); i += 16 {
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

	return db
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

func generateRandomIdKeyPairs(n int) map[string][]byte {
	// generate emails
	emails := generateRandomEmails(n)

	// generate key for each email and store in map
	out := make(map[string][]byte)
	for _, e := range emails {
		out[e] = generateRandom256BytesKey()
	}

	return out
}

func generateRandom256BytesKey() []byte {
	c := 256
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		panic("randomness generation error")
	}

	return b
}

func generateRandomEmails(n int) []string {
	emails := make([]string, n)
	for i := range emails {
		emails[i] = randomdata.Email()
	}

	return emails
}

func hashFunction(h maphash.Hash, dbLength uint64, id string) int {
	h.WriteString(id)
	sum := h.Sum64()
	h.Reset()

	return int(sum % dbLength)
}
