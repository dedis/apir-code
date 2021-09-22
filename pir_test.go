package main

// Test suite for classical PIR, used as baseline for the experiments.

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestPIRPointOneMb(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength * field.Bytes
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	// functions defined in vpir_test.go
	xofDB := utils.RandomPRG()
	xof := utils.RandomPRG()

	db := database.CreateRandomBytes(xofDB, dbLen, nRows, blockLen)

	retrievePIRPoint(t, xof, db, numBlocks, "PIRPointOneMb")
}

func retrievePIRPoint(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
	c := client.NewPIR(rnd, &db.Info)
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		in := make([]byte, 4)
		binary.BigEndian.PutUint32(in, uint32(i))
		queries, err := c.QueryBytes(in, 2)
		require.NoError(t, err)

		a0, err := s0.AnswerBytes(queries[0])
		require.NoError(t, err)
		a1, err := s1.AnswerBytes(queries[1])
		require.NoError(t, err)

		answers := [][]byte{a0, a1}

		res, err := c.ReconstructBytes(answers)
		require.NoError(t, err)
		require.Equal(t, db.Entries[i*db.BlockSize:(i+1)*db.BlockSize], res)
	}
	fmt.Printf("TotalCPU time %s: %.2fms\n", testName, totalTimer.Record())
}
