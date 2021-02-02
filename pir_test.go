package main

import (
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/stretchr/testify/require"
)

func TestMultiBitVectorOneMbBytes(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemBitSize := field.Bytes * 8
	nRows := 1
	nCols := dbLen / (elemBitSize * blockLen * nRows)

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksBytes(t, xof, db, nRows*nCols, "MultiBitVectorOneKbBytes")
}

func TestMultiBitMatrixOneMbBytes(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemBitSize := field.Bytes * 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksBytes(t, xof, db, numBlocks, "MultiBitMatrixOneKbBytes")
}

func retrieveBlocksBytes(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
	c := client.NewPIR(rnd, &db.Info)
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		queries := c.Query(i, 2)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])

		answers := [][]byte{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		require.Equal(t, db.Entries[i/db.NumColumns][(i%db.NumColumns)*db.BlockSize:(i%db.NumColumns+1)*db.BlockSize], res)

	}
	fmt.Printf("Total time %s: %.2fms\n", testName, totalTimer.Record())
}
