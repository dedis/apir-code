package pgp

import (
	"bytes"
	"crypto/x509"
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

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/nikirill/go-crypto/openpgp/packet"
)

const (
	eightKiB                = 8192
	sksParsedOutputFileName = "sks-dump.pgp"
	keySizeLimit            = eightKiB
)

// Key defines a PGP item after processing and saving into a binary file
type Key struct {
	Id     string
	Packet []byte
}

func AnalyzeDumpFiles(files []string) (map[string]*openpgp.Entity, error) {
	// map for the parsed entityMap
	entityMap := make(map[string]*openpgp.Entity)

	for _, file := range files {
		in, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		el, err := openpgp.ReadKeyRing(in)
		if err != nil {
			return nil, err
		}
		//var expired bool
		var email string
		for _, e := range el {
			// skip revoked keys
			if len(e.Revocations) > 0 {
				continue
			}
			email = PrimaryEmail(e)
			// skip keys without emails
			if email == "" {
				continue
			}
			// TODO: Should we skip expired keys?
			//expired, email = isExpired(e)
			//if expired {
			//	numExpired += 1
			//	continue
			//}

			//Remove subkeys (as a PoC) so that only the primary key is left
			e.Subkeys = nil
			// we index the entityMap by the primary identity and keep only
			// the latest key if there are multiple for a given identity
			if prev, ok := entityMap[email]; !ok {
				entityMap[email] = e
			} else {
				// save the entity if the primary key is fresher than the stored one
				if prev.PrimaryKey.CreationTime.Before(e.PrimaryKey.CreationTime) {
					entityMap[email] = e
				}
			}
		}
		if err = in.Close(); err != nil {
			log.Printf("Unable to close file %s", file)
			return nil, err
		}
	}

	return entityMap, nil
}

func WriteKeysOnDisk(dir string, entities map[string]*openpgp.Entity) error {
	var err error
	var buf bytes.Buffer
	// If the file already exists, the content is overwritten
	out, err := os.OpenFile(filepath.Join(dir, sksParsedOutputFileName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
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
		if err = encoder.Encode(&Key{Id: email, Packet: buf.Bytes()}); err != nil {
			return err
		}
		buf.Reset()
	}
	if err = out.Close(); err != nil {
		return err
	}

	return nil
}

func LoadKeysFromDisk(dir string) ([]*Key, error) {
	var err error
	keys := make([]*Key, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		f, err := os.Open(filepath.Join(dir, file.Name()))
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

// isExpired checks whether there is an user id with non-expired self-signature
// and replies with an email corresponding to the primary id or the non-expired Id.
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
		if strings.ToLower(e.PrimaryIdentity().UserId.Email) == email {
			return e, nil
		}
	}
	fmt.Println(email)
	fmt.Println(hex.EncodeToString(block))
	return nil, errors.New("no key with the given email id is found")
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

func writePublicKeysOnDisk(keys map[string][]byte) error {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	if err := e.Encode(keys); err != nil {
		return err
	}

	err := ioutil.WriteFile("keys.data", b.Bytes(), 0644)
	if err != nil {
		return err
	}

	// change permission
	// TODO: we really need this?
	err = os.Chmod("keys.data", 0644)
	if err != nil {
		return err
	}

	return nil
}

func marshalPublicKeys(primaryKeys map[string]*packet.PublicKey) map[string][]byte {
	m := make(map[string][]byte)
	for e, pk := range primaryKeys {
		//  MarshalPKIXPublicKey converts a public key to PKIX, ASN.1
		//  DER form. The encoded public key is a SubjectPublicKeyInfo
		//  structure (see RFC 5280, Section 4.1).
		// The following key types are currently supported: *rsa.PublicKey,
		// *ecdsa.PublicKey and ed25519.PublicKey. Unsupported key types result in an
		// error.
		// TODO: find a way to marshal *dsa.PublicKey
		b, err := x509.MarshalPKIXPublicKey(pk.PublicKey)
		if err != nil {
			fmt.Println("unsupported key", err)
			continue
		}
		m[e] = b
	}

	return m
}

func extractPrimaryKeys(el openpgp.EntityList) map[string]*packet.PublicKey {
	m := make(map[string]*packet.PublicKey)
	for _, e := range el {
		ids := ""
		for _, id := range e.Identities {
			ids += id.UserId.Email
		}
		m[ids] = e.PrimaryKey

	}

	return m
}

func importSingleDump(path string) (openpgp.EntityList, error) {
	// open single dump file
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	// read the keys
	el, err := openpgp.ReadKeyRing(f)
	if err != nil {
		return nil, err
	}

	return el, nil
}
