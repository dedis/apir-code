package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateRandomDB(t *testing.T) {
	path := "../../data/random_id_key.csv"
	db, err := GenerateRandomDB(path)
	require.NoError(t, err)

	fmt.Println(db)
}
