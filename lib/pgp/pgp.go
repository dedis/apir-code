package pgp

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nikirill/go-crypto/openpgp/armor"

	"github.com/nikirill/go-crypto/openpgp"
)

const (
	eightKiB              = 8192
	keySizeLimit          = eightKiB
	SksParsedFullFileName = "sks-full.pgp"
	SksOriginalFolder     = "sks-original"
	SksParsedFolder       = "sks"
)

// Key defines a PGP item after processing and saving into a binary file
type Key struct {
	ID     string
	Packet []byte
}

func AnalyzeKeyDump(files []string) (map[string]*openpgp.Entity, error) {
	// map for the parsed keys
	keys := make(map[string]*openpgp.Entity)

	for _, file := range files {
		fmt.Printf("Processing %s\n", file)
		in, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		el, err := openpgp.ReadKeyRing(in)
		if err != nil {
			return nil, err
		}
		for _, e := range el {
			saveKeyIfValid(e, keys)
		}
		if err = in.Close(); err != nil {
			log.Printf("Unable to close file %s\n", file)
			return nil, err
		}
	}

	return keys, nil
}

// Analyzes whether a given key is valid for us and, if so, saves it to the key map
func saveKeyIfValid(e *openpgp.Entity, keyMap map[string]*openpgp.Entity) {
	//var expired bool

	// skip revoked keys
	if len(e.Revocations) > 0 {
		return
	}
	email := PrimaryEmail(e)
	// skip keys without any email info
	if email == "" {
		return
	}

	//expired, email = isExpired(e)
	//if expired {
	//	return
	//}

	// remove subkeys (as a PoC) so that only the primary key is left
	e.Subkeys = nil
	// we index the keyMap by the primary identity and keep only
	// the latest key if there are multiple for a given identity
	if prev, ok := keyMap[email]; !ok {
		keyMap[email] = e
	} else {
		// save the entity if the primary key is fresher than the stored one
		if prev.PrimaryKey.CreationTime.Before(e.PrimaryKey.CreationTime) {
			keyMap[email] = e
		}
	}
}

func WriteKeysOnDisk(dir string, entities map[string]*openpgp.Entity) error {
	var err error
	var buf bytes.Buffer

	fmt.Printf("Saving to %s\n", filepath.Join(dir, SksParsedFullFileName))
	// If the file already exists, the content is overwritten
	out, err := os.OpenFile(filepath.Join(dir, SksParsedFullFileName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(out)

	for email, entity := range entities {
		err = entity.Serialize(&buf)
		if err != nil {
			return err
		}
		// If the serialized key packet is larger than an enforced upper-bound,
		// we ignore this key.
		if buf.Len() > keySizeLimit {
			buf.Reset()
			continue
		}
		if err = encoder.Encode(&Key{ID: email, Packet: buf.Bytes()}); err != nil {
			return err
		}
		buf.Reset()
	}
	if err = out.Close(); err != nil {
		return err
	}

	return nil
}

// Reads the given directory and returns the file names
// matching the sks dump description
func GetSksOriginalDumpFiles(dir string) ([]string, error) {
	sksRgx := `sks-dump-[0-9]{4}\.pgp`
	//sksRgx := `sks-dump-0000.pgp`
	return GetFilesThatMatch(dir, sksRgx)
}

// GetAllFiles returns all the filenames from the directory
func GetAllFiles(dir string) ([]string, error) {
	allFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0)
	for _, file := range allFiles {
		files = append(files, filepath.Join(dir, file.Name()))
	}
	return files, nil
}

// GetFilesThatMatch returns the filenames from the directory which
// match with the given regex expression
func GetFilesThatMatch(dir string, rgx string) ([]string, error) {
	allFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(rgx)
	files := make([]string, 0)
	for _, file := range allFiles {
		if re.MatchString(file.Name()) {
			files = append(files, filepath.Join(dir, file.Name()))
		}
	}
	return files, nil
}

func LoadKeysFromDisk(files []string) ([]*Key, error) {
	keys := make([]*Key, 0)
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		decoder := gob.NewDecoder(f)
		for {
			key := new(Key)
			// Decoding the serialized data
			if err = decoder.Decode(key); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			keys = append(keys, key)
		}
		if err = f.Close(); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

func LoadAndParseKeys(files []string) ([]*openpgp.Entity, error) {
	var entities openpgp.EntityList
	keys, err := LoadKeysFromDisk(files)
	if err != nil {
		return nil, err
	}
	el := make([]*openpgp.Entity, len(keys))
	for i, key := range keys {
		entities, err = openpgp.ReadKeyRing(bytes.NewBuffer(key.Packet))
		if len(entities) != 1 {
			return nil, errors.New("too many encoded entities")
		}
		el[i] = entities[0]
	}

	return el, nil
}

// isExpired checks whether there is an user id with non-expired self-signature
// and replies with an email corresponding to the primary id or the non-expired ID.
func isExpired(e *openpgp.Entity) (expired bool, email string) {
	expired = true
	for _, id := range e.Identities {
		if !id.SelfSignature.KeyExpired(time.Now()) {
			expired = false
			email = id.UserId.Email
			break
		}
	}
	if !e.PrimaryIdentity().SelfSignature.KeyExpired(time.Now()) && !expired {
		email = e.PrimaryIdentity().UserId.Email
	}

	return
}

// Returns the lower-cased email from the primary identity, or
// if it is empty, the alphabetically first non-empty lower-cased email
func PrimaryEmail(e *openpgp.Entity) string {
	email := e.PrimaryIdentity().UserId.Email
	// iterate over identities in search for the email if
	// the primary identity does not have one
	if email == "" {
		for _, id := range e.Identities {
			if id.UserId.Email != "" && (email == "" || email > id.UserId.Email) {
				email = id.UserId.Email
			}
		}
	}
	return strings.ToLower(email)
}

// Returns an Entity with the given email in the primary ID from a block of
// serialized entities.
func RecoverKeyFromBlock(block []byte, email string) (*openpgp.Entity, error) {
	// parse the input bytes as a key ring
	reader := bytes.NewReader(block)
	el, err := openpgp.ReadKeyRing(reader)
	if err != nil {
		return nil, err
	}
	// go over PGP entities and find the key with the given email as one of the ids
	for _, e := range el {
		if PrimaryEmail(e) == email {
			return e, nil
		}
	}
	log.Printf("The key with user email %s is not the block %s\n", email, hex.EncodeToString(block))
	return nil, errors.New("no key with the given email id is found")
}

func ArmorKey(entity *openpgp.Entity) (string, error) {
	var err error
	buf := new(bytes.Buffer)
	headers := map[string]string{"Comment": "Retrieved with Authenticated PIR"}
	arm, err := armor.Encode(buf, openpgp.PublicKeyType, headers)
	if err != nil {
		return "", err
	}
	err = entity.Serialize(arm)
	if err != nil {
		return "", err
	}
	if err = arm.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
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
	}
	return "", errors.New("email not found in the id")
}

// Regex for finding an email address surrounded by <>
func compileRegexToMatchEmail() *regexp.Regexp {
	email := `([a-zA-Z0-9_+\.-]+)@([a-zA-Z0-9\.-]+)\.([a-zA-Z\.]{2,10})`
	return regexp.MustCompile(`\<` + email + `\>`)
}
