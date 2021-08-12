package main

// Test suite for Merkle tree-based VPIR schemes. Only multi-bit schemes are
// implemented using this approach.

import (
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/stretchr/testify/require"
)

func TestMultiBitVectorOneMbMerkle(t *testing.T) {
	dbLen := oneMB // specified in bits
	// blockLen is computed such that a block contains the same amount of
	// bytes as in the integrated approach, which uses the finite filed in
	// /lib/field.
	blockLen := testBlockLength * constants.Bytes
	nRows := 1

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitMerkle(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksITMerkle(t, xof, db, db.NumRows*db.NumColumns, "DPFMultiBitVectorMerkle")
}

func TestMultiBitMatrixOneMbMerkle(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength * constants.Bytes
	// since this scheme works on bytes, the bit size of one element is 8
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitMerkle(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksITMerkle(t, xof, db, numBlocks, "MultiBitMatrixOneMbMerkle")
}

func TestDPFMultiBitVectorMerkle(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength * constants.Bytes
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nRows := 1

	// getXof is defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitMerkle(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksDPFMerkle(t, xof, db, numBlocks, "DPFMultiBitVectorMerkle")
}

func retrieveBlocksITMerkle(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
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
		require.NoError(t, err)
		require.Equal(t, db.Entries[i*db.BlockSize:(i+1)*db.BlockSize-db.ProofLen-1], res)
	}

	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}

func retrieveBlocksDPFMerkle(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
	c := client.NewPIRdpf(rnd, &db.Info)
	s0 := server.NewPIRdpf(db)
	s1 := server.NewPIRdpf(db)

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
		require.Equal(t, db.Entries[i*db.BlockSize:(i+1)*db.BlockSize-db.ProofLen-1], res)
	}

	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}
