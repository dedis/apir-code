package database

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateRealKeysDB(t *testing.T) {
	_, err := GenerateRealKeysDB(1)
	require.NoError(t, err)
}

func TestGetEmailAddressFromId(t *testing.T) {
	var email string
	var err error
	re := compileRegexToMatchEmail()
	// expected format
	email, err = getEmailAddressFromId("Alice Wonderland <alice@wonderland.com>", re)
	require.NoError(t, err)
	require.Equal(t, "alice@wonderland.com", email)

	// id without email
	email, err = getEmailAddressFromId("Alice Wonderland", re)
	require.Error(t, err)

	// empty email
	email, err = getEmailAddressFromId("Alice Wonderland <>", re)
	require.Error(t, err)

	// non-valid email
	email, err = getEmailAddressFromId("Bob <??@bob.bob>", re)
	require.Error(t, err)
}
