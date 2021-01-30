package database

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	numKeysToDBLengthRatio float32 = 0.2
	sixteenKiB                     = 16384
)

// Encoded PGP key in the sks json dump
type key struct {
	Id        []string `json:"id"`
	Packet    string   `json:"packet"`
	Timestamp int      `json:"timestamp"`
}

func GenerateRealKeyDB(dumpPath string, numRows, chunkLength int) (*DB, error) {
	var keys []*key
	var err error
	keys, err = ReadSKSKeyDump(dumpPath)
	if err != nil {
		return nil, err
	}
	// decide on the length of the hash table
	tableLen := int(float32(len(keys)) * numKeysToDBLengthRatio)
	ht, err := makeHashTable(keys, tableLen)

	// get the maximum byte length of the values in the hashTable
	// +1 takes into account the padding 0x80 that is always added.
	//maxBytes := utils.MaxBytesLength(ht) + 1
	maxBytes := utils.MaxBytesLength(ht)
	fmt.Println(maxBytes)
	blockLen := int(math.Ceil(float64(maxBytes)/float64(chunkLength)))
	numColumns := tableLen

	// create all zeros db
	db := CreateZeroMultiBitDB(numRows, numColumns, blockLen)

	// embed data into field elements
	for k, v := range ht {
		elements := field.ZeroVector(blockLen)
		// Pad the block
		//v = PadBlock(v)
		// embed all bytes
		for j := 0; j < len(v); j += chunkLength {
			end := j + chunkLength
			if end > len(v) {
				end = len(v)
			}
			e := new(field.Element).SetBytes(v[j:end])
			elements[j/chunkLength] = *e
		}
		// store in db last block and automatically pad since we start
		// with an all zeros db
		copy(db.Entries[k/numColumns][(k%numColumns)*blockLen:(k%numColumns+1)*blockLen], elements)
	}

	return db, nil
}

// Reads in the SKS pgp key dump in the JSON format
func ReadSKSKeyDump(dir string) ([]*key, error) {
	var err error
	keys := make([]*key, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		f, err := os.Open(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(f)
		for {
			var k key
			if err := decoder.Decode(&k); err == io.EOF {
				break
			} else if err != nil {
				return keys, err
			}
			keys = append(keys, &k)
		}
		if err = f.Close(); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

func makeHashTable(keys []*key, tableLen int) (map[int][]byte, error) {
	var err error
	var email string
	var packet []byte
	var seen bool

	// prepare db
	db := make(map[int][]byte)
	seenEmails := make(map[string]bool)

	// compile email regex pattern
	re := compileRegexToMatchEmail()

	// range over all id,v pairs and assign every pair to a given bucket
	for _, key := range keys {
		packet, err = hex.DecodeString(key.Packet)
		if err != nil {
			log.Printf("Problem decoding the pgp packet %s", key.Packet)
		}
		// There is handful of abnormally large keys, including a few larger
		// than 200 KiB. We filter them out.
		// TODO: Decide on the max size to accept. Fixed limit or compute distribution
		// TODO: and cut off 0.1% of the outliers?
		if len(key.Packet) > sixteenKiB {
			continue
		}
		// iterate over all the ids of the key and hash for each id separately
		// (the key is duplicated the length of the id times)
		for _, id := range key.Id {
			email, err = getEmailAddressFromPGPId(id, re)
			if err != nil {
				// if the id does not include an email, move on
				//log.Printf("id without email: %s", id)
				continue
			}
			// At the moment, we add a key for a given email only once.
			// If an email appears with multiple keys, only the first is added.
			// TODO: choose a logic for handling multiple keys per identity
			if _, seen = seenEmails[email]; !seen{
				seenEmails[email] = true
			} else {
				continue
			}
			hashKey := HashToIndex(email, tableLen)
			db[hashKey] = append(db[hashKey], packet...)
		}
	}

	return db, nil
}

// The PGP key ID typically has the form "Firstname Lastname <email address>".
// getEmailAddressFromPGPId parses the ID string and returns the email if found,
// or returns an empty string and an error otherwise.
func getEmailAddressFromPGPId(id string, re *regexp.Regexp) (string, error) {
	email := re.FindString(id)
	if email != "" {
		email = strings.Trim(email, "<")
		email = strings.Trim(email, ">")
		return email, nil
	} else {
		return "", errors.New("email not found in the id")
	}
}

// Regex for finding an email address surrounded by <>
func compileRegexToMatchEmail() *regexp.Regexp {
	email := `([a-zA-Z0-9_+\.-]+)@([a-zA-Z0-9\.-]+)\.([a-zA-Z\.]{2,10})`
	return regexp.MustCompile(`\<` + email + `\>`)
}

// Simple ISO/IEC 7816-4 padding where 0x80 is appended to the block and
// zeros are assumed afterwards
func PadBlock(block []byte) []byte{
	return append(block, byte(0x80))
}

func UnPadBlock(block []byte) []byte {
	// remove zeros
	block = bytes.TrimRightFunc(block, func(b rune) bool {
		return b == 0
	})
	// remove 0x80 preceding zeros
	return block[:len(block)-1]
}

func getNTopValuesFromMap(m map[string]int, n int) {
	// Turning the map into this structure
	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	// Then sorting the slice by value, higher first.
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	// Print the x top values
	for _, kv := range ss[:n] {
		fmt.Printf("%s, %d\n", kv.Key, kv.Value)
	}
}
