package main

// Test suite for the PoC application with real PGP keys. All the tests run
// locally, the networking logic is not tested here.

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

// TestRetrieveRealKeysDPFVector tests the retrieval of real PGP keys using
// the DPF-based multi-bit scheme. With DPF, the database is always represented
// as a vector.
func TestRetrieveRealKeysDPFVector(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	db, err := database.GenerateRealKeyDB(filePaths, constants.ChunkBytesLength, false)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	// client and servers
	prg := utils.RandomPRG()
	c := client.NewDPF(prg, &db.Info)
	servers := makeDPFServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

// TestRetrieveRealKeysITMatrix tests the retrieval of real PGP keys using the
// matrix-based multi-bit scheme.
func TestRetrieveRealKeysITMatrix(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	db, err := database.GenerateRealKeyDB(filePaths, constants.ChunkBytesLength, true)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	// client and servers
	prg := utils.RandomPRG()
	c := client.NewIT(prg, &db.Info)
	servers := makeITServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

// TestRetrieveRealKeysPIRDPFVector tests the retrieval of real PGP keys using
// the classical PIR DPF-based scheme. With DPF, the database is always
// represented as a vector.
func TestRetrieveRealKeysPIRDPFVector(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	db, err := database.GenerateRealKeyBytes(filePaths, false)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	// client and servers
	prg := utils.RandomPRG()
	c := client.NewPIRdpf(prg, &db.Info)
	servers := makePIRDPFServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

// TestRetrieveRealKeysPIRITMatrix tests the retrieval of real PGP keys using
// the classical PIR matrix-based scheme.
func TestRetrieveRealKeysPIRITMatrix(t *testing.T) {
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
	servers := makePIRITServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

// TestRetrieveRealKeysMerkleDPFVector tests the retrieval of real PGP keys using
// the classical PIR DPF-based scheme. With DPF, the database is always
// represented as a vector.
func TestRetrieveRealKeysMerkleDPFVector(t *testing.T) {
	// math randomness used only for testing purposes
	rand.Seed(time.Now().UnixNano())

	// get file paths for key dump
	filePaths := getDBFilePaths()

	// generate db from sks key dump
	db, err := database.GenerateRealKeyMerkle(filePaths, false)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	// client and servers
	prg := utils.RandomPRG()
	c := client.NewPIRdpf(prg, &db.Info)
	servers := makePIRDPFServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

// TestRetrieveRealKeysMerkleITMatrix tests the retrieval of real PGP keys using
// the classical PIR matrix-based scheme.
func TestRetrieveRealKeysMerkleITMatrix(t *testing.T) {
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
	servers := makePIRITServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

func retrieveRealKeyBlocks(t *testing.T, c client.Client, servers []server.Server, realKeys []*openpgp.Entity, numBlocks int) {
	// number of keys to retrieve for the test
	numKeys := 100

	start := time.Now()
	for i := 0; i < numKeys; i++ {
		j := rand.Intn(len(realKeys))
		result := retrieveBlockGivenID(t, c, servers, pgp.PrimaryEmail(realKeys[j]), numBlocks)
		//result = database.UnPadBlock(result)
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

	// query given hash key
	queries, err := c.QueryBytes(hashKey, len(ss))
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
	case []field.Element:
		// TODO: improve this
		padded := field.VectorToBytesAll(result.([]field.Element))
		padded = database.UnPadBlock(padded)
		bits := utils.BytesTobits(padded)
		return utils.BytesUnpadded(bits)
	default:
		return result.([]byte)
	}
}

func makeDPFServers(db *database.DB) []server.Server {
	s0 := server.NewDPF(db)
	s1 := server.NewDPF(db)
	return []server.Server{s0, s1}
}

func makeITServers(db *database.DB) []server.Server {
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)
	return []server.Server{s0, s1}
}

func makePIRDPFServers(db *database.Bytes) []server.Server {
	s0 := server.NewPIRdpf(db)
	s1 := server.NewPIRdpf(db)
	return []server.Server{s0, s1}
}

func makePIRITServers(db *database.Bytes) []server.Server {
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)
	return []server.Server{s0, s1}
}

func getDBFilePaths() []string {
	sksDir := filepath.Join("data", pgp.SksParsedFolder)
	// get a random chunk of the key dump in the folder
	//filePath := filepath.Join(sksDir, fmt.Sprintf("sks-%03d.pgp", rand.Intn(31)))
	filePath := filepath.Join(sksDir, "sks-022.pgp")
	fmt.Printf("Testing with %s\n", filePath)
	return []string{filePath}
}
