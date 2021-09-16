package main

// Test suite for the PoC application with real PGP keys. All the tests run
// locally, the networking logic is not tested here.

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

func TestRealCountEmailMatch(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	db, err := getDB()
	require.NoError(t, err)

	match := db.KeysInfo[rand.Intn(db.NumColumns)].UserId.Email
	h := blake2b.Sum256([]byte(match))
	in := utils.ByteToBits(h[:16])
	q := &query.ClientFSS{
		Target: query.UserId,
		Input:  in,
	}

	retrieveKeysFSS(t, db, q, match, "TestRealCountEntireEmail")
}

func TestRealCountStartsWithEmail(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	db, err := getDB()
	require.NoError(t, err)

	match := "bryan"
	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Target:    query.UserId,
		FromStart: len(match),
		Input:     in,
	}

	retrieveKeysFSS(t, db, q, match, "TestRealCountStartsWithEmail")
}

func TestRealCountEndsWithEmail(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	db, err := getDB()
	require.NoError(t, err)

	match := "com"
	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Target:  query.UserId,
		FromEnd: len(match),
		Input:   in,
	}

	retrieveKeysFSS(t, db, q, match, "TestRealCountEndsWithEmail")
}

func TestRealCountPublicKeyAlgorithm(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	db, err := getDB()
	require.NoError(t, err)

	match := packet.PubKeyAlgoRSA
	in := utils.ByteToBits([]byte{byte(match)})
	q := &query.ClientFSS{
		Target: query.PubKeyAlgo,
		Input:  in,
	}

	retrieveKeysFSS(t, db, q, match, "TestRealCountPublicKeyAlgorithm")
}

func retrieveKeysFSS(t *testing.T, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
	rnd := utils.RandomPRG()

	c := client.NewFSS(rnd, &db.Info)
	s0 := server.NewFSS(db, 0, c.Fss.PrfKeys)
	s1 := server.NewFSS(db, 1, c.Fss.PrfKeys)

	rand.Seed(time.Now().UnixNano())
	totalTimer := monitor.NewMonitor()

	// compute the input of the query
	fssKeys, err := c.QueryBytes(q, 2)
	require.NoError(t, err)

	a0, err := s0.AnswerBytes(fssKeys[0])
	require.NoError(t, err)
	a1, err := s1.AnswerBytes(fssKeys[1])
	require.NoError(t, err)

	answers := [][]byte{a0, a1}

	res, err := c.ReconstructBytes(answers)
	require.NoError(t, err)
	totalTime := totalTimer.Record()
	fmt.Printf("TotalCPU time %s: %.1fms, %.1fs\n", testName, totalTime, totalTime/float64(1000))
	// verify output
	count := uint32(0)
	for _, k := range db.KeysInfo {
		switch q.Target {
		case query.UserId:
			toMatch := ""
			email := k.UserId.Email
			if q.FromStart != 0 {
				if q.FromStart > len(email) {
					continue
				}
				toMatch = email[:q.FromStart]
			} else if q.FromEnd != 0 {
				if q.FromEnd > len(email) {
					continue
				}
				toMatch = email[len(email)-q.FromEnd:]
			} else {
				toMatch = email
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
	require.Equal(t, count, res.([]uint32)[0])
}

func getDB() (*database.DB, error) {
	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	return database.GenerateRealKeyDB(filePaths)
}

func getDBFilePaths() []string {
	sksDir := filepath.Join("data", pgp.SksParsedFolder)
	// get a random chunk of the key dump in the folder
	//filePath := filepath.Join(sksDir, fmt.Sprintf("sks-%03d.pgp", rand.Intn(31)))
	// filePaths := make([]string, 0)
	// for i := 0; i < 3; i++ {
	// 	fp := filepath.Join(sksDir, fmt.Sprintf("sks-%03d.pgp", i))
	// 	filePaths = append(filePaths, fp)
	// }
	// fmt.Println("Testing with", filePaths)
	// return filePaths
	filePath := filepath.Join(sksDir, "sks-000.pgp")
	fmt.Printf("Testing with %s\n", filePath)
	return []string{filePath}
}
