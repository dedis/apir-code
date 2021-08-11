package main

// Test suite for integrated VPIR.

import (
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/si-co/vpir-code/lib/constants"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

const (
	oneB            = 8
	oneKB           = 1024 * oneB
	oneMB           = 1024 * oneKB
	testBlockLength = 64
)

func TestMultiBitVectorOneMbVPIR(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength
	elemBitSize := constants.Bytes * 8
	nRows := 1
	nCols := dbLen / (elemBitSize * blockLen * nRows)

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db, err := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)
	require.NoError(t, err)

	retrieveBlocks(t, xof, db, nRows*nCols, "MultiBitVectorOneMbVPIR")
}

func TestSingleBitVectorOneKbVPIR(t *testing.T) {
	dbLen := oneKB
	nRows := 1
	nCols := dbLen

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db, err := database.CreateRandomSingleBitDB(xofDB, dbLen, nRows)
	require.NoError(t, err)

	retrieveBlocks(t, xof, db, nRows*nCols, "SingleBitVectorOneKbVPIR")
}

func TestMultiBitMatrixOneMbVPIR(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength
	elemBitSize := constants.Bytes * 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db, err := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)
	require.NoError(t, err)

	retrieveBlocks(t, xof, db, numBlocks, "MultiBitMatrixOneMbVPIR")
}

/*
func TestFSSMultiBitVectorVPIR(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength
	elemSize := 128
	numBlocks := dbLen / (elemSize * blockLen)
	nRows := 1

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")
	db, err := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)
	require.NoError(t, err)

	retrieveBlocksFSS(t, xof, db, numBlocks, "FSSMultiBitVectorVPIR")
}

func TestFSSMultiBitMatrixVPIR(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength
	elemSize := 128
	numBlocks := dbLen / (elemSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db, err := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)
	require.NoError(t, err)

	retrieveBlocksFSS(t, xof, db, numBlocks, "FSSMultiBitMatrixVPIR")
}
*/

func getXof(t *testing.T, key string) io.Reader {
	return utils.RandomPRG()
}

func retrieveBlocks(t *testing.T, rnd io.Reader, db *database.DB, numBlocks int, testName string) {
	c := client.NewIT(rnd, &db.Info)
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)

	totalTimer := monitor.NewMonitor()
	var elems []uint32
	for i := 0; i < numBlocks; i++ {
		queries := c.Query(i, 2)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])

		answers := [][]uint32{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		if db.BlockSize == constants.SingleBitBlockLength {
			elems = db.Range(i, i+1)
		} else {
			elems = db.Range(i*db.BlockSize, (i+1)*db.BlockSize)
		}
		require.Equal(t, elems, res)
	}
	fmt.Printf("TotalCPU time %s: %.2fms\n", testName, totalTimer.Record())
}

/*
func retrieveBlocksFSS(t *testing.T, rnd io.Reader, db *database.DB, numBlocks int, testName string) {
	c := client.NewFSS(rnd, &db.Info)
	s0 := server.NewFSS(db)
	s1 := server.NewFSS(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		fssKeys := c.Query(i, 2)

		a0 := s0.Answer(fssKeys[0])
		a1 := s1.Answer(fssKeys[1])

		answers := [][]uint32{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)

		elems := db.Range(i*db.BlockSize, (i+1)*db.BlockSize)

		require.Equal(t, elems, res)
	}

	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}
*/
