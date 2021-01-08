package gpg

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFromDisk(t *testing.T) {
	keys, err := ReadPublicKeysFromDisk()
	require.NoError(t, err)
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
