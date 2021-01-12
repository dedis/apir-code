package utils

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

// HashToIndex hashes the given id to an index for a database of the given
// length
func HashToIndex(id string, length int) int {
	hash := blake2b.Sum256([]byte(id))

	return int(binary.BigEndian.Uint64(hash[:])) % length
}

// AnalyzeIDKeys analyzes the given id->key samples and returns the maximal id
// bytes length and the maximal key bytes length, both expressed in byte
func AnalyzeIDKeys(in map[string][]byte) (int, int) {
	maxIDLength := 0
	maxKeyLength := 0

	for id, key := range in {
		idBytesLength := len([]byte(id))
		keyBytesLength := len(key)

		if maxIDLength < idBytesLength {
			maxIDLength = idBytesLength
		}

		if maxKeyLength < keyBytesLength {
			maxKeyLength = keyBytesLength
		}
	}

	return maxIDLength, maxKeyLength
}

// ParseCSVRandomIDKeys parse a csv file containing random id and keys and
// returns a map where keys are the random ids and values are the randomly
// generated cryptographic keys
func ParseCSVRandomIDKeys(path string) (map[string][]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	out := make(map[string][]byte)

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		key, err := base64.StdEncoding.DecodeString(record[1])
		if err != nil {
			return nil, err
		}

		out[record[0]] = key
	}

	return out, nil
}
