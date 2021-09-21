package main

// Test suite for integrated VPIR.

import (
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

const (
	oneB            = 8
	oneKB           = 1024 * oneB
	oneMB           = 1024 * oneKB
	testBlockLength = 64
	numIdentifiers  = 1000
)

func TestCountEntireEmail(t *testing.T) {
	match := "epflepflepflepflepflepflepflepfl"

	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomKeysDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		db.KeysInfo[i].UserId.Email = match
	}

	h := blake2b.Sum256([]byte(match))
	in := utils.ByteToBits(h[:16])
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId},
		Input: in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountEntireEmail")
}

func TestCountStartsWithEmail(t *testing.T) {
	match := "START"
	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomKeysDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		newEmail := match + db.KeysInfo[i].UserId.Email[5:]
		db.KeysInfo[i].UserId.Email = newEmail
	}

	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromStart: len(match)},
		Input: in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountStartsWithEmail")
}

func TestCountEndsWithEmail(t *testing.T) {
	match := "END"
	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomKeysDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		newEmail := db.KeysInfo[i].UserId.Email[:len(db.KeysInfo[i].UserId.Email)-len(match)] + match
		db.KeysInfo[i].UserId.Email = newEmail
	}

	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromEnd: len(match)},
		Input: in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountEndsWithEmail")
}

func TestCountPublicKeyAlgorithm(t *testing.T) {
	match := packet.PubKeyAlgoRSA
	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomKeysDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		db.KeysInfo[i].PubKeyAlgo = match
	}

	in := utils.ByteToBits([]byte{byte(match)})
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.PubKeyAlgo},
		Input: in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountPublicKeyAlgorithm")
}

func TestCountCreationTime(t *testing.T) {
	match := time.Date(2009, time.November, 0, 0, 0, 0, 0, time.UTC)
	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomKeysDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		db.KeysInfo[i].CreationTime = match
	}

	binaryMatch, err := match.MarshalBinary()
	require.NoError(t, err)
	in := utils.ByteToBits(binaryMatch)
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.CreationTime},
		Input: in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCreationDate")
}

func TestCountAndQuery(t *testing.T) {
	match := []interface{}{time.Date(2009, time.November, 0, 0, 0, 0, 0, time.UTC), packet.PubKeyAlgoRSA}
	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomKeysDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		db.KeysInfo[i].CreationTime = match[0].(time.Time)
		db.KeysInfo[i].PubKeyAlgo = match[1].(packet.PublicKeyAlgorithm)
	}

	//matchByte:= append([]byte(match[0].(string)), byte(match[1].(packet.PublicKeyAlgorithm)))
	matchBytes, err := match[0].(time.Time).MarshalBinary()
	require.NoError(t, err)
	matchBytes = append(matchBytes, byte(match[1].(packet.PublicKeyAlgorithm)))
	in := utils.ByteToBits(matchBytes)
	q := &query.ClientFSS{
		Info: &query.Info{
			And:     true,
			Targets: []query.Target{query.PubKeyAlgo, query.CreationTime},
		},
		Input: in,
	}

	retrieveBlocksFSS(t, xof, db, q, match, "TestCountAndQuery")
}

func retrieveBlocksFSS(t *testing.T, rnd io.Reader, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
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

	// verify output
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
		case query.PubKeyAlgo:
			if k.PubKeyAlgo == match {
				count++
			}
		case query.CreationTime:
			if k.CreationTime.Equal(match.(time.Time)) {
				count++
			}
		default:
			panic("unknown query type")
		}
	}

	// verify result
	require.Equal(t, count, res)
}
