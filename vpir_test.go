package main

// Test suite for integrated VPIR.

import (
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
	numIdentifiers  = 100
)

func TestCountEntireEmail(t *testing.T) {
	match := "epflepflepflepflepflepflepflepfl"

	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		db.KeysInfo[i].UserId.Email = match
	}

	in := utils.ByteToBits([]byte(match))
	q := query.ClientFSS{
		Target: query.UserId,
		Input:  in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountEntireEmail")
}

func TestCountStartsWithEmail(t *testing.T) {
	match := "START"
	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		newEmail := match + db.KeysInfo[i].UserId.Email[5:]
		db.KeysInfo[i].UserId.Email = newEmail
	}

	in := utils.ByteToBits([]byte(match))
	q := query.ClientFSS{
		Target:    query.UserId,
		FromStart: len(match),
		Input:     in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountStartsWithEmail")
}

func TestCountEndsWithEmail(t *testing.T) {
	match := "END"
	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		newEmail := db.KeysInfo[i].UserId.Email[:len(db.KeysInfo[i].UserId.Email)-3] + match
		db.KeysInfo[i].UserId.Email = newEmail
	}

	in := utils.ByteToBits([]byte(match))
	q := query.ClientFSS{
		Target:  query.UserId,
		FromEnd: len(match),
		Input:   in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountStartsWithEmail")
}

func retrieveBlocksFSS(t *testing.T, rnd io.Reader, db *database.DB, q query.ClientFSS, match, testName string) {
	c := client.NewFSS(rnd, &db.Info)
	s0 := server.NewFSS(db, 0, c.Fss.PrfKeys)
	s1 := server.NewFSS(db, 1, c.Fss.PrfKeys)

	rand.Seed(time.Now().UnixNano())
	totalTimer := monitor.NewMonitor()

	// compute the input of the query
	fssKeys := c.Query(q, 2)

	a0 := s0.Answer(fssKeys[0])
	a1 := s1.Answer(fssKeys[1])

	answers := [][]uint32{a0, a1}

	res, err := c.Reconstruct(answers)
	require.NoError(t, err)
	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())

	count := uint32(0)
	for _, k := range db.KeysInfo {
		switch q.Target {
		case query.UserId:
			toMatch := ""
			if q.FromStart != 0 {
				toMatch = k.UserId.Email[:q.FromStart]
			} else if q.FromEnd != 0 {
				toMatch = k.UserId.Email[len(k.UserId.Email)-q.FromEnd:]
			} else {
				toMatch = k.UserId.Email
			}

			if toMatch == match {
				count++
			}
		}
	}

	// verify result
	require.Equal(t, count, res[0])
}

func getXof(t *testing.T, key string) io.Reader {
	return utils.RandomPRG()
}
