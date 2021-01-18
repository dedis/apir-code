package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"github.com/si-co/vpir-code/lib/utils"
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
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

const (
	oneMB = 1048576 * 8
	oneKB = 1024 * 8
)


func TestRetrieveRandomKeyBlock(t *testing.T) {
	path := "data/random_id_key.csv"

	// generate db from data
	db, idLength, keyLength, blockLength, err := database.GenerateRandomDB(path)
	require.NoError(t, err)

	var key1 utils.PRGKey
	copy(key1[:], []byte("my key"))
	prg := utils.NewPRG(&key1)

	// client and servers
	c := client.NewDPF(prg)
	s0 := server.NewDPF(db)
	s1 := server.NewDPF(db)

	// open id->key file
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	// Parse the file
	r := csv.NewReader(f)

	// scheme variables
	chunkLength := constants.ChunkBytesLength

	// helping variables
	zeroSlice := make([]byte, 45)

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
		hashKey := utils.HashToIndex(expectedID, constants.DBLength)

		// query given hash key
		fssKeys := c.Query(hashKey, 2)

		// get servers answers
		a0 := s0.Answer(fssKeys[0], 0)
		a1 := s1.Answer(fssKeys[1], 1)
		answers := [][][]field.Element{a0, a1}

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

		require.Equal(t, expectedKey, idKey[expectedID])
		fmt.Printf("Total time retrieve key: %.1fms\n", totalTimer.Record())

		// retrieve only one key
		break
	}
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

		answers := [][][]field.Element{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		require.ElementsMatch(t, db.Entries[i/db.NumColumns][i%db.NumColumns], res)
	}
	fmt.Printf("Total time %s: %.2fms\n", testName, totalTimer.Record())
}

func getXof(t *testing.T, key string) io.Reader {
	xof, err := blake2b.NewXOF(0, []byte(key))
	require.NoError(t, err)
	return xof
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

	retrieveBlocks(t, xof, db, nRows*nCols,"MultiBitVectorOneKb")
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

/*func TestMatrixOneKbByte(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiMatrixOneKbByte()
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	require.NoError(t, err)

	rebalanced := true
	c := client.NewITSingleByte(xof, rebalanced)
	s0 := server.NewITSingleByte(rebalanced, db)
	s1 := server.NewITSingleByte(rebalanced, db)
	s2 := server.NewITSingleByte(rebalanced, db)
	for i := 0; i < 8191; i++ {
		queries := c.Query(i, 3)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])
		a2 := s2.Answer(queries[2])

		answers := [][]byte{a0, a1, a2}

		_, err := c.Reconstruct(answers)
		require.NoError(t, err)
	}
	fmt.Printf("Total time MatrixOneKbByte: %.1fms\n", totalTimer.Record())
}

func TestVectorByte(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiVectorByte()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	rebalanced := false
	c := client.NewITSingleByte(xof, rebalanced)
	s0 := server.NewITSingleByte(rebalanced, db)
	s1 := server.NewITSingleByte(rebalanced, db)
	s2 := server.NewITSingleByte(rebalanced, db)
	m := monitor.NewMonitor()
	for i := 0; i < 136; i++ {
		m.Reset()
		queries := c.Query(i, 3)
		fmt.Printf("Query: %.3fms\t", m.RecordAndReset())

		a0 := s0.Answer(queries[0])
		fmt.Printf("Answer 1: %.3fms\t", m.RecordAndReset())

		a1 := s1.Answer(queries[1])
		fmt.Printf("Answer 2: %.3fms\t", m.RecordAndReset())

		a2 := s2.Answer(queries[2])
		fmt.Printf("Answer 3: %.3fms\t", m.RecordAndReset())

		answers := [][]byte{a0, a1, a2}

		m.Reset()
		x, err := c.Reconstruct(answers)
		fmt.Println(x)
		require.NoError(t, err)
		fmt.Printf("Reconstruct: %.3fms\n", m.RecordAndReset())
		if x == byte(0) {
			result += "0"
		} else {
			result += "1"
		}
	}

	testBitResult(t, result)

	fmt.Printf("Total time VectorByte: %.1fms\n", totalTimer.Record())
}*/

func TestDPFMulti(t *testing.T) {
	dbLenMB := 1048576 * 8
	xofDB, err := blake2b.NewXOF(0, []byte("db key"))
	require.NoError(t, err)
	db := database.CreateRandomMultiBitOneMBGF(xofDB, dbLenMB, constants.BlockLength)

	xof, err := blake2b.NewXOF(0, []byte("my key"))
	require.NoError(t, err)

	totalTimer := monitor.NewMonitor()

	c := client.NewDPF(xof)
	s0 := server.NewDPF(db)
	s1 := server.NewDPF(db)

	fieldElements := 128 * 8

	for i := 0; i < fieldElements/16; i++ {
		fssKeys := c.Query(i, constants.BlockLength, 2)

		a0 := s0.Answer(fssKeys[0], 0, constants.BlockLength)
		a1 := s1.Answer(fssKeys[1], 1, constants.BlockLength)

		answers := [][]field.Element{a0, a1}

		res, err := c.Reconstruct(answers, constants.BlockLength)
		require.NoError(t, err)
		require.ElementsMatch(t, db.Entries[i], res)
	}

	fmt.Printf("Total time dpf-based MultiBitOneKb: %.1fms\n", totalTimer.Record())
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
*/

func testBitResult(t *testing.T, result string) {
	b, err := utils.BitStringToBytes(result)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	expected := "Playing with VPIR"
	output := string(b)
	if expected != output {
		t.Errorf("Expected '%v' but got '%v'", expected, output)
	}
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
