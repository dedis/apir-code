package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromKeysFile(t *testing.T) {
	_, err := FromKeysFile()
	require.NoError(t, err)
}
