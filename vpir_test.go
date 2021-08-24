package main

// Test suite for integrated VPIR.

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
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

func TestMultiBitVPIR(t *testing.T) {
	keyToDownload := 50
	numIdentifiers := 100

	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomDB(rndDB, numIdentifiers)
	require.NoError(t, err)

	retrieveBlocksFSS(t, xof, db, keyToDownload, "FSSMultiBitVectorVPIR")
}

func retrieveBlocksFSS(t *testing.T, rnd io.Reader, db *database.DB, numBlocks int, testName string) {
	c := client.NewFSS(rnd, &db.Info)
	s0 := server.NewFSS(db, 0, c.Fss.PrfKeys)
	s1 := server.NewFSS(db, 1, c.Fss.PrfKeys)

	rand.Seed(time.Now().UnixNano())
	totalTimer := monitor.NewMonitor()
	var j int
	for i := 0; i < numBlocks; i++ {
		// get random identifier number
		j = rand.Intn(db.NumColumns)

		// compute corresponding identifier
		id := int(binary.BigEndian.Uint32(db.Identifiers[j*constants.IdentifierLength : (j+1)*constants.IdentifierLength]))

		fssKeys := c.Query(id, 2)

		a0 := s0.Answer(fssKeys[0])
		a1 := s1.Answer(fssKeys[1])

		answers := [][]uint32{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)

		elems := db.Range(j*db.BlockSize, (j+1)*db.BlockSize)

		require.Equal(t, elems, res)
	}

	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}

func getXof(t *testing.T, key string) io.Reader {
	return utils.RandomPRG()
}
