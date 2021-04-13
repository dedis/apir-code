package pgp

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/stretchr/testify/require"
)

func TestSerialization(t *testing.T) {
	var m map[string]*openpgp.Entity
	var entities openpgp.EntityList
	var err error
	var buf bytes.Buffer

	m, _ = analyzeRandomSksDumpFile(t)
	for _, key := range m {
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
	tmpDir := "../../data/sks-tmp/"

	m1, fileName := analyzeRandomSksDumpFile(t)
	err = WriteKeysOnDisk(tmpDir, m1)
	require.NoError(t, err)

	m2, err = LoadKeysFromDisk([]string{fileName})
	for _, key := range m2 {
		//fmt.Printf("%s\n", key.ID)
		entities, err = openpgp.ReadKeyRing(bytes.NewBuffer(key.Packet))
		require.NoError(t, err)
		require.Equal(t, 1, len(entities))
		require.Equal(t, m1[key.ID].PrimaryKey, entities[0].PrimaryKey)
		require.Equal(t, PrimaryEmail(m1[key.ID]), PrimaryEmail(entities[0]))
		require.Equal(t, key.ID, PrimaryEmail(m1[key.ID]))
	}
}

func analyzeRandomSksDumpFile(t *testing.T) (map[string]*openpgp.Entity, string) {
	files, err := ioutil.ReadDir("../../data/sks-dump/")
	require.NoError(t, err)
	j := rand.Intn(len(files))

	m, err := AnalyzeKeyDump([]string{files[j].Name()})
	require.NoError(t, err)
	return m, files[j].Name()
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

func TestLoadKeysFromDisk(t *testing.T) {
	sksDir := filepath.Join("../../data", SksParsedFolder)
	files, err := GetFilesThatMatch(sksDir, SksParsedFullFileName)
	require.NoError(t, err)
	keys, err := LoadKeysFromDisk(files)
	require.NoError(t, err)
	log.Println(len(keys))
}