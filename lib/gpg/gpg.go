package main

import (
	"os"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func extractPrimaryKeys(el openpgp.EntityList) map[string]*packet.PublicKey {
	m := make(map[string]*packet.PublicKey)
	for _, e := range el {
		ids := ""
		for name := range e.Identities {
			ids += name
		}
		m[ids] = e.PrimaryKey

	}

	return m
}

func importSingleDump(path string) (openpgp.EntityList, error) {
	// open single dump file
	f, err := os.Open(path)
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
