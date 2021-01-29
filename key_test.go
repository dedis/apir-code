package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
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
	db, err := database.GenerateKeyDB(path, chunkLength, nRows, nCols)
	require.NoError(t, err)

	idLength := db.IDLength
	keyLength := db.KeyLength

	prg := utils.RandomPRG()

	// client and servers
	c := client.NewDPF(prg, db.Info)
	s0 := server.NewDPF(db, 0)
	s1 := server.NewDPF(db, 1)

	// open id->key file
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	// Parse the file
	r := csv.NewReader(f)

	// helping variables
	zeroSlice := make([]byte, idLength)

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

		// compute hash key for id
		hashKey := database.HashToIndex(expectedID, nCols*nRows)

		// query given hash key
		fssKeys := c.Query(hashKey, 2)

		//fmt.Println("client 0", "query:", fssKeys[0])
		//fmt.Println("client 1", "query:", fssKeys[1])

		// get servers answers
		a0 := s0.Answer(fssKeys[0])
		a1 := s1.Answer(fssKeys[1])
		answers := [][]field.Element{a0, a1}

		//fmt.Println("client 0", "answer:", answers[0])
		//fmt.Println("client 1", "answer:", answers[1])

		// reconstruct block
		result, err := c.Reconstruct(answers)
		require.NoError(t, err)

		// retrieve result bytes
		resultBytes := field.VectorToBytes(result)

		lastElementBytes := keyLength % chunkLength
		keyLengthWithPadding := int(math.Ceil(float64(keyLength)/float64(chunkLength))) * chunkLength
		totalLength := idLength + keyLengthWithPadding

		// parse block entries
		idKey := make(map[string]string)
		for i := 0; i < len(resultBytes)-totalLength+1; i += totalLength {
			idBytes := resultBytes[i : i+idLength]
			// test if we are in padding elements already
			if bytes.Equal(idBytes, zeroSlice) {
				break
			}
			idReconstructed := string(bytes.Trim(idBytes, "\x00"))

			keyBytes := resultBytes[i+idLength : i+idLength+keyLengthWithPadding]
			// remove padding for last element
			if lastElementBytes != 0 {
				keyBytes = append(keyBytes[:len(keyBytes)-chunkLength], keyBytes[len(keyBytes)-(lastElementBytes):]...)
			}

			// encode key
			idKey[idReconstructed] = base64.StdEncoding.EncodeToString(keyBytes)
		}

		//fmt.Println(idKey[expectedID])
		require.Equal(t, expectedKey, idKey[expectedID])

		// retrieve only one key
		//break
	}
	fmt.Printf("Total time retrieve key: %.1fms\n", totalTimer.Record())
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
