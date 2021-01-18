package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateRandomDB(t *testing.T) {
	path := "../../data/random_id_key.csv"
	_, _, _, err := GenerateRandomDB(path)
	require.NoError(t, err)
}
