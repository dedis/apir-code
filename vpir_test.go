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
	"golang.org/x/crypto/blake2b"
)

const (
	oneMB = 1048576 * 8
	oneKB = 1024 * 8
)

func TestMultiBitVectorOneMb(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemBitSize := field.Bytes * 8
	nRows := 1
	nCols := dbLen / (elemBitSize * blockLen * nRows)

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)

	retrieveBlocks(t, xof, db, nRows*nCols, "MultiBitVectorOneMb")
}

func TestSingleBitVectorOneKb(t *testing.T) {
	dbLen := oneKB
	nRows := 1
	nCols := dbLen

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomSingleBitDB(xofDB, dbLen, nRows)

	retrieveBlocks(t, xof, db, nRows*nCols, "SingleBitVectorOneMb")
}

func TestMultiBitMatrixOneMb(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemBitSize := field.Bytes * 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)
	retrieveBlocks(t, xof, db, numBlocks, "MultiBitMatrixOneKb")
}

func TestSingleBitMatrixOneKb(t *testing.T) {
	dbLen := oneKB - 92 // making the length a square
	numBlocks := dbLen
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomSingleBitDB(xofDB, dbLen, nRows)

	retrieveBlocks(t, xof, db, numBlocks, "SingleBitMatrixOneKb")
}

func TestDPFMultiVector(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemSize := 128
	numBlocks := dbLen / (elemSize * blockLen)
	nRows := 1

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")
	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksDPF(t, xof, db, numBlocks, "TestDPFMultiVector")
}

func TestDPFMultiMatrix(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemSize := 128
	numBlocks := dbLen / (elemSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")
	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)
	retrieveBlocksDPF(t, xof, db, numBlocks, "TestDPFMultiMatrix")
}

func getXof(t *testing.T, key string) io.Reader {
	xof, err := blake2b.NewXOF(0, []byte(key))
	require.NoError(t, err)
	return xof
}

func retrieveBlocks(t *testing.T, rnd io.Reader, db *database.DB, numBlocks int, testName string) {
	c := client.NewIT(rnd, &db.Info)
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		queries := c.Query(i, 2)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])

		answers := [][]field.Element{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		if db.BlockSize == constants.SingleBitBlockLength {
			require.ElementsMatch(t, db.Entries[i/db.NumColumns][i%db.NumColumns:i%db.NumColumns+1], res)
		} else {
			require.ElementsMatch(t, db.Entries[i/db.NumColumns][(i%db.NumColumns)*db.BlockSize:(i%db.NumColumns+1)*db.BlockSize], res)
		}

	}
	fmt.Printf("Total time %s: %.2fms\n", testName, totalTimer.Record())
}

func retrieveBlocksDPF(t *testing.T, rnd io.Reader, db *database.DB, numBlocks int, testName string) {
	c := client.NewDPF(rnd, &db.Info)
	s0 := server.NewDPF(db, 0)
	s1 := server.NewDPF(db, 1)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		fssKeys := c.Query(i, 2)

		a0 := s0.Answer(fssKeys[0])
		a1 := s1.Answer(fssKeys[1])

		answers := [][]field.Element{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		require.ElementsMatch(t, db.Entries[i/db.NumColumns][(i%db.NumColumns)*db.BlockSize:(i%db.NumColumns+1)*db.BlockSize], res)
	}

	fmt.Printf("Total time dpf-based %s: %.1fms\n", testName, totalTimer.Record())
}
