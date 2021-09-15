package main

// Test suite for the PoC application with real PGP keys. All the tests run
// locally, the networking logic is not tested here.

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

// TestRetrieveRealKeysDPFVector tests the retrieval of real PGP keys using
// the DPF-based multi-bit scheme. With DPF, the database is always represented
// as a vector.

func TestRealCountEmailMatch(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	db, err := database.GenerateRealKeyDB(filePaths)
	require.NoError(t, err)

	//// read in the real pgp key values
	//realKeys, err := pgp.LoadAndParseKeys(filePaths)
	//require.NoError(t, err)

	match := "arthurthompson@google.com"
	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Target: query.UserId,
		Input:  in,
	}


	retrieveKeysFSS(t, db, q, match, "TestCountEntireEmail")
}

func retrieveKeysFSS(t *testing.T, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
	rnd := utils.RandomPRG()

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
	require.Equal(t, count, res[0])
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
