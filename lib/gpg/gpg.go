package main

import (
	"bytes"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func readPublicKeysFromDisk() (map[string][]byte, error) {
	b, err := ioutil.ReadFile("keys.data")
	if err != nil {
		return nil, err
	}

	var keys map[string][]byte
	d := gob.NewDecoder(bytes.NewReader(b))

	// Decoding the serialized data
	if err = d.Decode(&keys); err != nil {
		return nil, err
	}

	return keys, nil
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
