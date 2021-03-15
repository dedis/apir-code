package main

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

func TestRetrieveRealKeysDPFVector(t *testing.T) {
	filePaths := getDBFilePaths()
	// Generate db from sks key dump
	db, err := database.GenerateRealKeyDB(filePaths, constants.ChunkBytesLength, false)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	prg := utils.RandomPRG()
	// client and servers
	c := client.NewDPF(prg, &db.Info)
	servers := makeDPFServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

func TestRetrieveRealKeysITMatrix(t *testing.T) {
	filePaths := getDBFilePaths()
	// Generate db from sks key dump
	db, err := database.GenerateRealKeyDB(filePaths, constants.ChunkBytesLength, true)
	require.NoError(t, err)
	numBlocks := db.NumColumns * db.NumRows

	// read in the real pgp key values
	realKeys, err := pgp.LoadAndParseKeys(filePaths)
	require.NoError(t, err)

	prg := utils.RandomPRG()
	// client and servers
	c := client.NewIT(prg, &db.Info)
	servers := makeITServers(db)

	retrieveRealKeyBlocks(t, c, servers, realKeys, numBlocks)
}

func retrieveRealKeyBlocks(t *testing.T, c client.Client, servers []server.Server, realKeys []*openpgp.Entity, numBlocks int) {
	var err error
	var j int
	var retrievedKey *openpgp.Entity

	numKeys := 100

	rand.Seed(time.Now().UnixNano())
	//totalTimer := monitor.NewMonitor()
	start := time.Now()
	for i := 0; i < numKeys; i++ {
		j = rand.Intn(len(realKeys))
		fmt.Println(pgp.PrimaryEmail(realKeys[j]))
		result := retrieveBlockGivenID(t, c, servers, pgp.PrimaryEmail(realKeys[j]), numBlocks)
		result = database.UnPadBlock(result)
		// Get a key from the block with the id of the search
		//retrievedKey, err = pgp.RecoverKeyFromBlock(result, pgp.PrimaryEmail(realKeys[j]))
		//require.NoError(t, err)
		//require.Equal(t, pgp.PrimaryEmail(realKeys[j]), pgp.PrimaryEmail(retrievedKey))
		//require.Equal(t, realKeys[j].PrimaryKey.Fingerprint, retrievedKey.PrimaryKey.Fingerprint)
	}
	//fmt.Printf("TotalCPU time to retrieve %d real keys: %.1fms\n", numKeys, totalTimer.Record())
	fmt.Printf("TotalCPU time to retrieve %d real keys: %v\n", numKeys, time.Since(start))
}

func retrieveBlockGivenID(t *testing.T, c client.Client, ss []server.Server, id string, dbLenBlocks int) []byte {
	var err error
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
	return field.VectorToBytes(result.([]field.Element))
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

func getDBFilePaths() []string {
	rand.Seed(time.Now().UnixNano())
	sksDir := filepath.Join("data", pgp.SksParsedFolder)
	// get a random chunk of the key dump in the folder
	filePath := filepath.Join(sksDir, fmt.Sprintf("sks-%03d.pgp", rand.Intn(31)))
	//filePath := filepath.Join(sksDir, "sks-022.pgp")
	fmt.Printf("Testing with %s\n", filePath)
	return []string{filePath}
}
