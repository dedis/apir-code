package database

import (
	"crypto/rand"
	"hash/maphash"

	"github.com/Pallinder/go-randomdata"
	"github.com/si-co/vpir-code/lib/constants"
)

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
