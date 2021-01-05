package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

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
