package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromKeysFile(t *testing.T) {
	db, err := FromKeysFile()
	require.NoError(t, err)

	fmt.Println(db)
}
