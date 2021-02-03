package pgp

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

//func TestLoadFromDisk(t *testing.T) {
//	_, err := LoadKeysFromDisk()
//	require.NoError(t, err)
//}

func TestImportEntireDump(t *testing.T) {
	errorCounts := 0
	basePath := ""
	if basePath == "" {
		panic("basicPath undefined")
	}

	for i := 0; i < 288; i++ {
		path := fmt.Sprintf(basePath+"/sks-dump-%04d.pgp", i)
		_, err := importSingleDump(path)
		if err != nil {
			errorCounts++
			log.Println(err, "with file: ", path)
		}
	}

	fmt.Println("Total errors: ", errorCounts, " over 287 files.")
}

func TestMarshalPublicKeysFromDump(t *testing.T) {
	path := ""
	if path == "" {
		panic("path not specified")
	}
	el, err := importSingleDump(path)
	require.NoError(t, err)

	primaryKeys := extractPrimaryKeys(el)

	keys := marshalPublicKeys(primaryKeys)

	err = writePublicKeysOnDisk(keys)
	require.NoError(t, err)
}

func TestImportSingleDump(t *testing.T) {
	path := ""
	if path == "" {
		panic("path not specified")
	}

	el, err := importSingleDump(path)
	require.NoError(t, err)

	primaryKeys := extractPrimaryKeys(el)

	for id, key := range primaryKeys {
		fmt.Println("id: ", id, " key: ", key)
	}
}

func TestGetEmailAddressFromId(t *testing.T) {
	var email string
	var err error
	re := compileRegexToMatchEmail()
	// expected format
	email, err = getEmailAddressFromPGPId("Alice Wonderland <alice@wonderland.com>", re)
	require.NoError(t, err)
	require.Equal(t, "alice@wonderland.com", email)

	// still valid email
	email, err = getEmailAddressFromPGPId("Michael Steiner <m1.steiner@von.ulm.de>", re)
	require.NoError(t, err)
	require.Equal(t, "m1.steiner@von.ulm.de", email)

	// id without email
	email, err = getEmailAddressFromPGPId("Alice Wonderland", re)
	require.Error(t, err)

	// empty email
	email, err = getEmailAddressFromPGPId("Alice Wonderland <>", re)
	require.Error(t, err)

	// non-valid email
	email, err = getEmailAddressFromPGPId("Bob <??@bob.bob>", re)
	require.Error(t, err)
}