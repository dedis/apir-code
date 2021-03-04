package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
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

func TestRetrieveRandomKeyBlockVector(t *testing.T) {
	// TODO: How do we choose dbLen (hence, nCols) ?
	dbLen := 40 * oneKB
	// maximum number of bytes embedded in a field elements
	chunkLength := constants.ChunkBytesLength
	nRows := 1
	nCols := dbLen / (nRows * chunkLength)
	retrieveRandomKeyBlock(t, chunkLength, nRows, nCols)
}

func TestRetrieveRandomKeyBlockMatrix(t *testing.T) {
	// TODO: How do we choose dbLen (hence, nCols) ?
	dbLen := 40 * oneKB
	// maximum number of bytes embedded in a field elements
	chunkLength := constants.ChunkBytesLength
	nRows := int(math.Sqrt(float64(dbLen / chunkLength)))
	nCols := nRows
	retrieveRandomKeyBlock(t, chunkLength, nRows, nCols)
}

func retrieveRealKeyBlocks(t *testing.T, c client.Client, servers []server.Server, realKeys []*openpgp.Entity, numBlocks int) {
	var err error
	var j int
	var retrievedKey *openpgp.Entity

	numKeys := 100

	rand.Seed(time.Now().UnixNano())
	totalTimer := monitor.NewMonitor()
	for i := 0; i < numKeys; i++ {
		j = rand.Intn(len(realKeys))
		fmt.Println(pgp.PrimaryEmail(realKeys[j]))
		result := retrieveBlockGivenID(t, c, servers, pgp.PrimaryEmail(realKeys[j]), numBlocks)
		result = database.UnPadBlock(result)
		// Get a key from the block with the id of the search
		retrievedKey, err = pgp.RecoverKeyFromBlock(result, pgp.PrimaryEmail(realKeys[j]))
		require.NoError(t, err)
		require.Equal(t, pgp.PrimaryEmail(realKeys[j]), pgp.PrimaryEmail(retrievedKey))
		require.Equal(t, realKeys[j].PrimaryKey.Fingerprint, retrievedKey.PrimaryKey.Fingerprint)
	}
	fmt.Printf("TotalCPU time to retrieve %d real keys: %.1fms\n", numKeys, totalTimer.Record())
}

func retrieveRandomKeyBlock(t *testing.T, chunkLength, nRows, nCols int) {
	path := "data/random_id_key.csv"

	// generate db from data
	db, err := database.GenerateRandomKeyDB(path, chunkLength, nRows, nCols)
	require.NoError(t, err)

	prg := utils.RandomPRG()

	// client and servers
	c := client.NewDPF(prg, &db.Info)
	servers := makeDPFServers(db)

	// open id->key file
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	// Parse the file
	r := csv.NewReader(f)

	totalTimer := monitor.NewMonitor()
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		// for testing
		expectedID := record[0]
		expectedKey := record[1]

		// retrieve result bytes
		result := retrieveBlockGivenID(t, c, servers, expectedID, nRows*nCols)
		validateRandomKey(t, expectedID, expectedKey, result, &db.Info, chunkLength)

		// retrieve only one key
		break
	}
	fmt.Printf("TotalCPU time retrieve key: %.1fms\n", totalTimer.Record())
}

func validateRandomKey(t *testing.T, id, key string, result []byte, dbInfo *database.Info, chunkLength int) (string, string) {
	var idReconstructed string
	var keyBytes []byte
	lastElementBytes := dbInfo.KeyLength % chunkLength
	keyLengthWithPadding := int(math.Ceil(float64(dbInfo.KeyLength)/float64(chunkLength))) * chunkLength
	totalLength := dbInfo.IDLength + keyLengthWithPadding

	// helping variables
	zeroSlice := make([]byte, dbInfo.IDLength)
	// parse block entries
	idKey := make(map[string]string)
	for i := 0; i < len(result)-totalLength+1; i += totalLength {
		idBytes := result[i : i+dbInfo.IDLength]
		// test if we are in padding elements already
		if bytes.Equal(idBytes, zeroSlice) {
			break
		}
		idReconstructed = string(bytes.Trim(idBytes, "\x00"))

		keyBytes = result[i+dbInfo.IDLength : i+dbInfo.IDLength+keyLengthWithPadding]
		// remove padding for last element
		if lastElementBytes != 0 {
			keyBytes = append(keyBytes[:len(keyBytes)-chunkLength], keyBytes[len(keyBytes)-(lastElementBytes):]...)
		}
		idKey[idReconstructed] = base64.StdEncoding.EncodeToString(keyBytes)
	}
	require.Equal(t, key, idKey[id])

	return idReconstructed, base64.StdEncoding.EncodeToString(keyBytes)
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
	//filePath := filepath.Join(sksDir, fmt.Sprintf("sks-%03d.pgp", rand.Intn(31)))
	filePath := filepath.Join(sksDir, "sks-026.pgp")
	fmt.Printf("Testing with %s\n", filePath)
	return []string{filePath}
}
