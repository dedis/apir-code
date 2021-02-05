package pgp

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/stretchr/testify/require"
)

func TestSerialization(t *testing.T) {
	var m1 map[string]*openpgp.Entity
	var entities openpgp.EntityList
	var err error
	var buf bytes.Buffer

	m1, err = AnalyzeDumpFiles([]string{"../../data/sks-dump/sks-dump-0000.pgp"})
	require.NoError(t, err)
	for _, key := range m1 {
		err = key.Serialize(&buf)
		require.NoError(t, err)
		entities, err = openpgp.ReadKeyRing(bytes.NewBuffer(buf.Bytes()))
		require.NoError(t, err)
		require.Equal(t, 1, len(entities))
		require.Equal(t, key.PrimaryKey.PublicKey, entities[0].PrimaryKey.PublicKey)
		buf.Reset()
	}
}

func TestWriteThenLoadKeys(t *testing.T) {
	var m1 map[string]*openpgp.Entity
	var m2 []*Key
	var entities openpgp.EntityList
	var err error
	sksDir := "../../data/sks/"

	m1, err = AnalyzeDumpFiles([]string{"../../data/sks-dump/sks-dump-0000.pgp"})
	require.NoError(t, err)
	err = WriteKeysOnDisk(sksDir, m1)
	require.NoError(t, err)
	m2, err = LoadKeysFromDisk(sksDir)
	for _, key := range m2 {
		//fmt.Printf("%s\n", key.Id)
		entities, err = openpgp.ReadKeyRing(bytes.NewBuffer(key.Packet))
		require.NoError(t, err)
		require.Equal(t, 1, len(entities))
		require.Equal(t, m1[key.Id].PrimaryKey, entities[0].PrimaryKey)
		require.Equal(t, PrimaryEmail(m1[key.Id]), PrimaryEmail(entities[0]))
		require.Equal(t, key.Id, PrimaryEmail(m1[key.Id]))
	}
}

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