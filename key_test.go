package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"github.com/si-co/vpir-code/lib/pgp"
	"golang.org/x/crypto/openpgp"
	"io"
	"math"
	"os"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestRetrieveRealKeysVector(t *testing.T) {
	var key *openpgp.Entity
	var err error

	sksPath := "data/sks/"
	numKeysToTest := 2
	nRows := 1
	// Generate db from sks key dump
	db, err := database.GenerateRealKeyDB(sksPath, nRows, constants.ChunkBytesLength)
	require.NoError(t, err)

	prg := utils.RandomPRG()

	// client and servers
	c := client.NewDPF(prg, &db.Info)
	s0 := server.NewDPF(db, 0)
	s1 := server.NewDPF(db, 1)
	servers := []*server.DPF{s0, s1}

	emails := []string{"carol@mail.com", "m1.steiner@von.ulm.de"}
	for i := 0; i < numKeysToTest; i++ {
		result := retrieveBlockGivenId(t, c, servers, emails[i], db.NumColumns*db.NumRows)
		key, err = pgp.RecoverKeyGivenEmail(result, emails[i])
		require.NoError(t, err)
		fmt.Println(key.PrimaryKey.PublicKey)
		fmt.Println(key.Identities)
	}

}

func TestRetrieveRandomKeyBlockVector(t *testing.T) {
	// TODO: How do we choose dbLen (hence, nCols) ?
	dbLen := 40 * oneKB
	// maximum numer of bytes embedded in a field elements
	chunkLength := constants.ChunkBytesLength
	nRows := 1
	nCols := dbLen / (nRows * chunkLength)
	retrieveRandomKeyBlock(t, chunkLength, nRows, nCols)
}

func TestRetrieveRandomKeyBlockMatrix(t *testing.T) {
	// TODO: How do we choose dbLen (hence, nCols) ?
	dbLen := 40 * oneKB
	// maximum numer of bytes embedded in a field elements
	chunkLength := constants.ChunkBytesLength
	nRows := int(math.Sqrt(float64(dbLen / chunkLength)))
	nCols := nRows
	retrieveRandomKeyBlock(t, chunkLength, nRows, nCols)
}

func retrieveRandomKeyBlock(t *testing.T, chunkLength, nRows, nCols int) {
	path := "data/random_id_key.csv"

	// generate db from data
	db, err := database.GenerateRandomKeyDB(path, chunkLength, nRows, nCols)
	require.NoError(t, err)

	prg := utils.RandomPRG()

	// client and servers
	c := client.NewDPF(prg, &db.Info)
	s0 := server.NewDPF(db, 0)
	s1 := server.NewDPF(db, 1)
	servers := []*server.DPF{s0, s1}

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
		result := retrieveBlockGivenId(t, c, servers, expectedID, nRows*nCols)
		validateRandomKey(t, expectedID, expectedKey, result, &db.Info, chunkLength)

		// retrieve only one key
		break
	}
	fmt.Printf("Total time retrieve key: %.1fms\n", totalTimer.Record())
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

func retrieveBlockGivenId(t *testing.T, c *client.DPF, ss []*server.DPF, id string, dbLenBlocks int) []byte {
	// compute hash key for id
	hashKey := database.HashToIndex(id, dbLenBlocks)

	// query given hash key
	fssKeys := c.Query(hashKey, len(ss))

	// get servers answers
	answers := make([][]field.Element, len(ss))
	for i := range ss {
		answers[i] = ss[i].Answer(fssKeys[i])
	}
	// reconstruct block
	result, err := c.Reconstruct(answers)
	require.NoError(t, err)

	// return result bytes
	return field.VectorToBytes(result)
}

/*
func TestRetrieveKey(t *testing.T) {
	db, err := database.FromKeysFile()
	require.NoError(t, err)
	blockLength := 40

	xof, err := blake2b.NewXOF(0, []byte("my key"))
	require.NoError(t, err)
	rebalanced := false

	c := client.NewITMulti(xof, rebalanced)
	s0 := server.NewITMulti(rebalanced, db)
	s1 := server.NewITMulti(rebalanced, db)

	for i := 0; i < 1; i++ {
		queries := c.Query(i, blockLength, 2)

		a0 := s0.Answer(queries[0], blockLength)
		a1 := s1.Answer(queries[1], blockLength)

		answers := [][]field.Element{a0, a1}

		result, err := c.Reconstruct(answers, blockLength)
		require.NoError(t, err)

		// parse result
		// TODO: logic for this should be in lib/gpg
		//lengthBytes := result[0].Bytes()
		//length, _ := binary.Varint(lengthBytes[len(lengthBytes)-1:])

		resultBytes := make([]byte, 0)
		for i := 0; i < len(result); i++ {
			elementBytes := result[i].Bytes()
			//fmt.Println("recon:", elementBytes)
			resultBytes = append(resultBytes, elementBytes[:]...)
		}
		elementsLength, _ := binary.Varint([]byte{resultBytes[0]})
		lastElementLength, _ := binary.Varint([]byte{resultBytes[1]})

		fmt.Println("")
		fmt.Println(elementsLength)
		fmt.Println(lastElementLength)
		fmt.Println(resultBytes[2 : 14+(elementsLength-2)*16+1])

		pub, err := x509.ParsePKIXPublicKey(resultBytes)
		if err != nil {
			log.Printf("failed to parse DER encoded public key: %v", err)
		} else {

			switch pub := pub.(type) {
			case *rsa.PublicKey:
				fmt.Println("pub is of type RSA:", pub)
			case *dsa.PublicKey:
				fmt.Println("pub is of type DSA:", pub)
			case *ecdsa.PublicKey:
				fmt.Println("pub is of type ECDSA:", pub)
			case ed25519.PublicKey:
				fmt.Println("pub is of type Ed25519:", pub)
			default:
				panic("unknown type of public key")
			}
		}
	}
}*/
