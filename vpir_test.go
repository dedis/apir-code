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
	"golang.org/x/crypto/blake2b"
)

const (
	oneMB = 1048576 * 8
	oneKB = 1024 * 8
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

	var key1 utils.PRGKey
	copy(key1[:], []byte("my key"))
	prg := utils.NewPRG(&key1)

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

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		totalTimer := monitor.NewMonitor()

		// for testing
		expectedID := record[0]
		expectedKey := record[1]

		// compute hash key for id
		hashKey := database.HashToIndex(expectedID, nCols*nRows)

		// query given hash key
		fssKeys := c.Query(hashKey, 2)

		// get servers answers
		a0 := s0.Answer(fssKeys[0])
		a1 := s1.Answer(fssKeys[1])
		answers := [][]field.Element{a0, a1}

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
		fmt.Printf("Total time retrieve key: %.1fms\n", totalTimer.Record())

		// retrieve only one key
		break
	}
}

func TestMultiBitVectorOneKb(t *testing.T) {
	dbLen := oneKB
	blockLen := constants.BlockLength
	elemSize := field.Bits
	nRows := 1
	nCols := dbLen / (elemSize * blockLen * nRows)

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)

	retrieveBlocks(t, xof, db, nRows*nCols, "MultiBitVectorOneKb")
}

func TestSingleBitVectorOneKb(t *testing.T) {
	dbLen := oneKB
	nRows := 1
	nCols := dbLen

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomSingleBitDB(xofDB, dbLen, nRows)

	retrieveBlocks(t, xof, db, nRows*nCols, "SingleBitVectorOneKb")
}

func TestMultiBitMatrixOneKb(t *testing.T) {
	dbLen := oneKB
	blockLen := constants.BlockLength
	elemSize := field.Bits
	numBlocks := dbLen / (elemSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)

	retrieveBlocks(t, xof, db, numBlocks, "MultiBitMatrixOneKb")
}

func TestSingleBitMatrixOneKb(t *testing.T) {
	dbLen := oneKB - 92 // making the length a square
	numBlocks := dbLen
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomSingleBitDB(xofDB, dbLen, nRows)

	retrieveBlocks(t, xof, db, numBlocks, "SingleBitMatrixOneKb")
}

func TestDPFMultiVector(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemSize := 128
	numBlocks := dbLen / (elemSize * blockLen)
	nRows := 1

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")
	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksDPF(t, xof, db, numBlocks, "TestDPFMultiVector")
}

func TestDPFMultiMatrix(t *testing.T) {
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemSize := 128
	numBlocks := dbLen / (elemSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")
	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)
	retrieveBlocksDPF(t, xof, db, numBlocks, "TestDPFMultiMatrix")
}

func getXof(t *testing.T, key string) io.Reader {
	xof, err := blake2b.NewXOF(0, []byte(key))
	require.NoError(t, err)
	return xof
}

func retrieveBlocks(t *testing.T, rnd io.Reader, db *database.DB, numBlocks int, testName string) {
	c := client.NewITClient(rnd, db.Info)
	s0 := server.NewITServer(db)
	s1 := server.NewITServer(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		queries := c.Query(i, 2)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])

		answers := [][]field.Element{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		require.ElementsMatch(t, db.Entries[i/db.NumColumns][i%db.NumColumns], res)
	}
	fmt.Printf("Total time %s: %.2fms\n", testName, totalTimer.Record())
}

func retrieveBlocksDPF(t *testing.T, rnd io.Reader, db *database.DB, numBlocks int, testName string) {
	c := client.NewDPF(rnd, db.Info)
	s0 := server.NewDPF(db, 0)
	s1 := server.NewDPF(db, 1)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		fssKeys := c.Query(i, 2)

		a0 := s0.Answer(fssKeys[0])
		a1 := s1.Answer(fssKeys[1])

		answers := [][]field.Element{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		require.ElementsMatch(t, db.Entries[i/db.NumColumns][i%db.NumColumns], res)
	}

	fmt.Printf("Total time dpf-based %s: %.1fms\n", testName, totalTimer.Record())
}

/*
func TestDPFSingle(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiVectorGF()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}

	blockLen := constants.SingleBitBlockLength

	c := client.NewDPF(xof)
	s0 := server.NewDPF(db)
	s1 := server.NewDPF(db)

	for i := 0; i < 136; i++ {
		fssKeys := c.Query(i, blockLen, 2)

		a0 := s0.Answer(fssKeys[0], 0, blockLen)
		a1 := s1.Answer(fssKeys[1], 1, blockLen)

		answers := [][]field.Element{a0, a1}

		x, err := c.Reconstruct(answers, blockLen)
		if err != nil {
			panic(err)
		}
		result += x[0].String()
	}

	testBitResult(t, result)

	fmt.Printf("Total time: %.1fms\n", totalTimer.Record())
}

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
