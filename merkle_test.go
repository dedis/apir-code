package main

// Test suite for Merkle tree-based VPIR schemes. Only multi-bit schemes are
// implemented using this approach.

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

func TestMerkle(t *testing.T) {
	numServers := 2
	dbLen := oneMB
	blockLen := testBlockLength * field.Bytes
	// since this scheme works on bytes, the bit size of one element is 8
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	db := database.CreateRandomMerkle(utils.RandomPRG(), dbLen, nRows, blockLen)

	retrieveBlocksMerkle(t, utils.RandomPRG(), db, numServers, numBlocks, "Merkle")
}

func TestMerkleFourServers(t *testing.T) {
	numServers := 4
	dbLen := oneMB
	blockLen := testBlockLength * field.Bytes
	// since this scheme works on bytes, the bit size of one element is 8
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	db := database.CreateRandomMerkle(utils.RandomPRG(), dbLen, nRows, blockLen)

	retrieveBlocksMerkle(t, utils.RandomPRG(), db, numServers, numBlocks, "MerkleFourServers")
}

func retrieveBlocksMerkle(t *testing.T, rnd io.Reader, db *database.Bytes, numServers, numBlocks int, testName string) {
	c := client.NewPIR(rnd, &db.Info)
	servers := make([]*server.PIR, numServers)
	for i := range servers {
		servers[i] = server.NewPIR(db)
	}

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		in := make([]byte, 4)
		binary.BigEndian.PutUint32(in, uint32(i))
		queries, err := c.QueryBytes(in, numServers)
		require.NoError(t, err)

		answers := make([][]byte, numServers)
		for i, s := range servers {
			a, err := s.AnswerBytes(queries[i])
			require.NoError(t, err)
			answers[i] = a
		}

		res, err := c.ReconstructBytes(answers)
		require.NoError(t, err)
		require.Equal(t, db.Entries[i*db.BlockSize:(i+1)*db.BlockSize-db.ProofLen-1], res)
	}

	fmt.Printf("Total CPU time %s: %.1fms\n", testName, totalTimer.Record())
}
