package main

// Test suite for classical PIR, used as baseline for the experiments.

import (
	"fmt"
	"io"
	"math"
	"runtime"
	"testing"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/stretchr/testify/require"
)

func TestMultiBitVectorOneMbPIR(t *testing.T) {
	dbLen := oneMB
	// we want to download the same number of bytes
	// as in the field representation
	blockLen := testBlockLength * field.Bytes
	elemBitSize := 8
	nRows := 1
	nCols := dbLen / (elemBitSize * blockLen * nRows)

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksBytes(t, xof, db, nRows*nCols, "MultiBitVectorOneMbPIR")
}

func TestMultiBitMatrixOneMbPIR(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength * field.Bytes
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksBytes(t, xof, db, numBlocks, "MultiBitMatrixOneMbPIR")
}

func TestDPFMultiBitVectorPIR(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength * field.Bytes
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nRows := 1

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")
	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksDPFBytes(t, xof, db, numBlocks, "DPFMultiBitVectorPIR")
}

func retrieveBlocksBytes(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
	c := client.NewPIR(rnd, &db.Info)
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		queries, err := c.QueryBytes(i, 2)
		require.NoError(t, err)

		a0, err := s0.AnswerBytes(queries[0])
		require.NoError(t, err)
		a1, err := s1.AnswerBytes(queries[1])
		require.NoError(t, err)

		answers := [][]byte{a0, a1}

		res, err := c.ReconstructBytes(answers)
		//fmt.Println(res)
		require.NoError(t, err)
		require.Equal(t, db.Entries[i*db.BlockSize:(i+1)*db.BlockSize], res)
	}
	fmt.Printf("TotalCPU time %s: %.2fms\n", testName, totalTimer.Record())
}

func retrieveBlocksDPFBytes(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
	c := client.NewPIRdpf(rnd, &db.Info)
	s0 := server.NewPIRdpf(db, 1)
	s1 := server.NewPIRdpf(db, 1)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		fssKeys, err := c.QueryBytes(i, 2)
		require.NoError(t, err)

		a0, err := s0.AnswerBytes(fssKeys[0])
		require.NoError(t, err)
		a1, err := s1.AnswerBytes(fssKeys[1])
		require.NoError(t, err)

		answers := [][]byte{a0, a1}

		res, err := c.ReconstructBytes(answers)
		require.NoError(t, err)
		require.Equal(t, db.Entries[i*db.BlockSize:(i+1)*db.BlockSize], res)
	}

	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}

func TestPIRNumberThreads(t *testing.T) {
	GiB := 1024 * 1024 * 1024 * 8
	dbLen := 3 * GiB
	blockLen := 8 * 1024
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)
	for cores := 1; cores <= runtime.NumCPU(); cores++ {
		c := client.NewPIR(xof, &db.Info)
		s0 := server.NewPIR(db, cores)
		s1 := server.NewPIR(db, cores)
		fmt.Printf("Cores: %d ", cores)

		clock := time.Now()
		queries, err := c.QueryBytes(88, 2)
		require.NoError(t, err)
		//fmt.Printf("query gen: %v  ", time.Since(clock))

		a0, err := s0.AnswerBytes(queries[0])
		require.NoError(t, err)
		a1, err := s1.AnswerBytes(queries[1])
		require.NoError(t, err)
		//fmt.Printf("two answers: %v  ", time.Since(clock))

		answers := [][]byte{a0, a1}

		_, err = c.ReconstructBytes(answers)
		require.NoError(t, err)
		fmt.Printf("Total time: %v  \n", time.Since(clock))
	}
}
