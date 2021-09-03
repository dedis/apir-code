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
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/query"
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
	numIdentifiers := 100

	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomDB(rndDB, numIdentifiers)
	require.NoError(t, err)

	retrieveBlocksFSS(t, xof, db, "FSSMultiBitVectorVPIR")
}

func retrieveBlocksFSS(t *testing.T, rnd io.Reader, db *database.DB, testName string) {
	c := client.NewFSS(rnd, &db.Info)
	s0 := server.NewFSS(db, 0, c.Fss.PrfKeys)
	s1 := server.NewFSS(db, 1, c.Fss.PrfKeys)

	rand.Seed(time.Now().UnixNano())
	totalTimer := monitor.NewMonitor()
	// compute corresponding identifier
	match := "epfl.edu"
	id := binary.BigEndian.Uint64([]byte(match))
	idBool := utils.ByteToBits([]byte(match))

	fssKeys := c.Query(int(id), idBool, query.UserId, 2)

	a0 := s0.Answer(fssKeys[0])
	a1 := s1.Answer(fssKeys[1])

	answers := [][]uint32{a0, a1}

	res, err := c.Reconstruct(answers)
	require.NoError(t, err)
	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())

	count := uint32(0)
	for _, k := range db.KeysInfo {
		if k.UserId.Email == match {
			count++
		}
	}

	// verify result
	require.Equal(t, count, res[0])
}

func getXof(t *testing.T, key string) io.Reader {
	return utils.RandomPRG()
}
