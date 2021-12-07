package main

// Test suite for the PoC application with real PGP keys. All the tests run
// locally, the networking logic is not tested here.

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

var db *database.DB

func initRealDB() {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// init global db
	var err error
	db, err = getDB()
	if err != nil {
		panic(err)
	}

	// GC after DB creation
	runtime.GC()
}

func TestRealRetrieveKey(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	db, err := database.GenerateRealKeyMerkle(filePaths, true)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	// client and servers
	prg := utils.RandomPRG()
	c := client.NewPIR(prg, &db.Info)
	servers := []server.Server{server.NewPIR(db), server.NewPIR(db)}

	retrieveRealKey(t, c, servers, realKeys, numBlocks)
}

func TestRealRetrieveKeyPIR(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	db, err := database.GenerateRealKeyBytes(filePaths, true)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	// client and servers
	prg := utils.RandomPRG()
	c := client.NewPIR(prg, &db.Info)
	servers := []server.Server{server.NewPIR(db), server.NewPIR(db)}

	retrieveRealKey(t, c, servers, realKeys, numBlocks)
}

func TestRealCountEmail(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := emailMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountEntireEmail")
}

func TestRealCountEmailPIR(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := emailMatch(db)
	retrieveComplexPIR(t, db, q, match, "TestRealCountEntireEmailPIR")
}

func TestRealCountStartsWithEmail(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := startsWithMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountStartsWithEmail")
}

func TestRealCountStartsWithEmailPIR(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := startsWithMatch(db)
	retrieveComplexPIR(t, db, q, match, "TestCountStartsWithEmailPIR")
}

func TestRealCountEndsWithEmail(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := endsWithMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountEndsWithEmail")
}

func TestRealCountEndsWithEmailPIR(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := endsWithMatch(db)
	retrieveComplexPIR(t, db, q, match, "TestRealCountEndsWithEmailPIR")
}

func TestRealCountPublicKeyAlgorithm(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := pkaMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountPublicKeyAlgorithm")
}

func TestRealCountPublicKeyAlgorithmPIR(t *testing.T) {
	if db == nil {
		initRealDB()
	}
	match, q := pkaMatch(db)
	retrieveComplex(t, db, q, match, "TestRealCountPublicKeyAlgorithmPIR")
}

func retrieveRealKey(t *testing.T, c client.Client, servers []server.Server, realKeys []*openpgp.Entity, numBlocks int) {
	// number of keys to retrieve for the test
	numKeys := 1

	start := time.Now()
	for i := 0; i < numKeys; i++ {
		// get random key
		j := rand.Intn(len(realKeys))
		fmt.Println(pgp.PrimaryEmail(realKeys[i]))
		result := retrieveBlockGivenID(t, c, servers, pgp.PrimaryEmail(realKeys[j]), numBlocks)
		result = database.UnPadBlock(result)

		// Get a key from the block with the id of the search
		retrievedKey, err := pgp.RecoverKeyFromBlock(result, pgp.PrimaryEmail(realKeys[j]))
		require.NoError(t, err)
		require.Equal(t, pgp.PrimaryEmail(realKeys[j]), pgp.PrimaryEmail(retrievedKey))
		require.Equal(t, realKeys[j].PrimaryKey.Fingerprint, retrievedKey.PrimaryKey.Fingerprint)
	}
	fmt.Printf("TotalCPU time to retrieve %d real keys: %v\n", numKeys, time.Since(start))
}

func retrieveBlockGivenID(t *testing.T, c client.Client, ss []server.Server, id string, dbLenBlocks int) []byte {
	// compute hash key for id
	hashKey := database.HashToIndex(id, dbLenBlocks)
	in := make([]byte, 4)
	binary.BigEndian.PutUint32(in, uint32(hashKey))

	// query given hash key
	queries, err := c.QueryBytes(in, len(ss))
	require.NoError(t, err)

	// get servers answers
	answers := make([][]byte, len(ss))
	for i := range ss {
		answers[i], err = ss[i].AnswerBytes(queries[i])
		require.NoError(t, err)

	}

	// reconstruct block
	result, err := c.ReconstructBytes(answers)
	require.NoError(t, err)

	// return result bytes
	switch result.(type) {
	case []uint32:
		return field.VectorToBytes(result.([]uint32))
	default:
		return result.([]byte)
	}
}

func retrieveComplexPIR(t *testing.T, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
	c := client.NewPIRfss(utils.RandomPRG(), &db.Info)
	s0 := server.NewPIRfss(db, 0)
	s1 := server.NewPIRfss(db, 1)

	totalTimer := monitor.NewMonitor()

	f, err := os.Create("complexPIR.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	// compute the input query
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
	count := localResult(db, q.Info, match)
	require.Equal(t, count, res.(uint32))
}

func retrieveComplex(t *testing.T, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
	c := client.NewFSS(utils.RandomPRG(), &db.Info)
	s0 := server.NewFSS(db, 0)
	s1 := server.NewFSS(db, 1)

	totalTimer := monitor.NewMonitor()

	f, err := os.Create("complexVPIR.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

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
	count := localResult(db, q.Info, match)
	require.Equal(t, count, res.(uint32))
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

func localResult(db *database.DB, q *query.Info, match interface{}) uint32 {
	count := uint32(0)
	for _, k := range db.KeysInfo {
		if !q.And {
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
		} else {
			email := k.UserId.Email
			if len(email) < q.FromEnd {
				continue
			}
			toMatchEmail := email[len(email)-q.FromEnd:]
			if k.CreationTime.Year() == (match.([]interface{}))[0].(time.Time).Year() &&
				toMatchEmail == (match.([]interface{}))[1].(string) {
				count++
			}
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
