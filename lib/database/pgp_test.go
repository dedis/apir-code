package database

import (
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateRealKeysDB(t *testing.T) {
	sksFilesPath := "../../data/sks/"
	nRows := 1
	_, err := GenerateRealKeyDB(sksFilesPath, nRows, constants.ChunkBytesLength)
	require.NoError(t, err)
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

func TestPadBlock(t *testing.T) {
	b := []byte{0x01, 0xff, 0x35}
	b = PadBlock(b)
	require.Equal(t, []byte{0x01, 0xff, 0x35, 0x80}, b)
}

func TestUnPadBlock(t *testing.T) {
	b := []byte{0x01, 0xff, 0x35, 0x80, 0x00, 0x00, 0x00}
	b = UnPadBlock(b)
	require.Equal(t, []byte{0x01, 0xff, 0x35}, b)
}
