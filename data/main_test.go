package main

import (
	"fmt"
	"testing"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	db, err := database.LoadDB("db", "vpir")
	require.NoError(t, err)

	fmt.Println(db.Info)
}
