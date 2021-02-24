package main

import (
	"testing"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

func TestMultiBitVectorOneMbMerkle(t *testing.T) {
	dbLen := oneMB
	// we want to download the same numer of bytes
	// as in the field representation
	blockLen := testBlockLength * field.Bytes
	elemBitSize := 8
	nRows := 1
	nCols := dbLen / (elemBitSize * blockLen * nRows)

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitMerle(xofDB, dbLen, nRows, blockLen)
}
