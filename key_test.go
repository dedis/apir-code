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

var db *database.DB

func init() {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// init global db
	var err error
	db, err = getDB()
	if err != nil {
		panic(err)
	}
}

func TestRealCountEmail(t *testing.T) {
	match, q := emailMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountEntireEmail")
}

func TestRealCountEmailPIR(t *testing.T) {
	match, q := emailMatch(db)
	retrieveComplexPIR(t, db, q, match, "TestRealCountEntireEmailPIR")
}

func TestRealCountStartsWithEmail(t *testing.T) {
	match, q := startsWithMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountStartsWithEmail")
}

func TestRealCountStartsWithEmailPIR(t *testing.T) {
	match, q := startsWithMatch(db)
	retrieveComplexPIR(t, db, q, match, "TestCountStartsWithEmailPIR")
}

func TestRealCountEndsWithEmail(t *testing.T) {
	match, q := endsWithMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountEndsWithEmail")
}

func TestRealCountEndsWithEmailPIR(t *testing.T) {
	match, q := endsWithMatch(db)
	retrieveComplexPIR(t, db, q, match, "TestRealCountEndsWithEmailPIR")
}

func TestRealCountPublicKeyAlgorithm(t *testing.T) {
	match, q := pkaMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountPublicKeyAlgorithm")
}

func TestRealCountPublicKeyAlgorithmPIR(t *testing.T) {
	match, q := pkaMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountPublicKeyAlgorithmPIR")
}

func retrieveComplexPIR(t *testing.T, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
	c := client.NewPIRfss(utils.RandomPRG(), &db.Info)
	s0 := server.NewPIRfss(db, 0, c.Fss.PrfKeys)
	s1 := server.NewPIRfss(db, 1, c.Fss.PrfKeys)

	totalTimer := monitor.NewMonitor()

	// compute the input query
	fssKeys := c.Query(q, 2)

	a0 := s0.Answer(fssKeys[0])
	a1 := s1.Answer(fssKeys[1])

	answers := []int{a0, a1}

	res, err := c.Reconstruct(answers)
	require.NoError(t, err)
	totalTime := totalTimer.Record()
	fmt.Printf("TotalCPU time %s: %.1fms, %.1fs\n", testName, totalTime, totalTime/float64(1000))

	// verify result
	count := localResult(db, q.Info, match)
	require.Equal(t, count, res)
}

func retrieveComplex(t *testing.T, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
	c := client.NewFSS(utils.RandomPRG(), &db.Info)
	s0 := server.NewFSS(db, 0, c.Fss.PrfKeys)
	s1 := server.NewFSS(db, 1, c.Fss.PrfKeys)

	totalTimer := monitor.NewMonitor()

	// compute the input of the query
	in, err := q.Encode()
	require.NoError(t, err)
	fssKeys, err := c.QueryBytes(in, 2)
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

	// verify result
	count := uint32(localResult(db, q.Info, match))
	require.Equal(t, count, res)
}

func emailMatch(db *database.DB) (string, *query.ClientFSS) {
	match := db.KeysInfo[rand.Intn(db.NumColumns)].UserId.Email
	h := blake2b.Sum256([]byte(match))
	in := utils.ByteToBits(h[:16])
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId},
		Input: in,
	}

	return match, q
}

func startsWithMatch(db *database.DB) (string, *query.ClientFSS) {
	email := db.KeysInfo[rand.Intn(db.NumColumns)].UserId.Email
	for len(email) < 5 {
		email = db.KeysInfo[rand.Intn(db.NumColumns)].UserId.Email
	}
	match := email[:5]
	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromStart: len(match)},
		Input: in,
	}

	return match, q
}

func endsWithMatch(db *database.DB) (string, *query.ClientFSS) {
	email := db.KeysInfo[rand.Intn(db.NumColumns)].UserId.Email
	for len(email) < 5 {
		email = db.KeysInfo[rand.Intn(db.NumColumns)].UserId.Email
	}
	match := email[len(email)-5:]
	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromEnd: len(match)},
		Input: in,
	}

	return match, q
}

func pkaMatch(db *database.DB) (packet.PublicKeyAlgorithm, *query.ClientFSS) {
	match := packet.PubKeyAlgoRSA
	in := utils.ByteToBits([]byte{byte(match)})
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.PubKeyAlgo},
		Input: in,
	}

	return match, q
}

func localResult(db *database.DB, q *query.Info, match interface{}) int {
	count := 0
	for _, k := range db.KeysInfo {
		switch q.Target {
		case query.UserId:
			toMatch := ""
			if q.FromStart != 0 {
				email := k.UserId.Email
				if len(email) < q.FromStart {
					continue
				}
				toMatch = k.UserId.Email[:q.FromStart]
			} else if q.FromEnd != 0 {
				email := k.UserId.Email
				if len(email) < q.FromEnd {
					continue
				}
				toMatch = email[len(email)-q.FromEnd:]
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

	return count
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
