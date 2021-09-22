package main

// Test suite for classical PIR, used as baseline for the experiments.

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"testing"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

func TestPIRComplex(t *testing.T) {
	match := "epflepflepflepflepflepflepflepfl"

	rndDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db, err := database.CreateRandomKeysDB(rndDB, numIdentifiers)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		db.KeysInfo[i].UserId.Email = match
	}

	h := blake2b.Sum256([]byte(match))
	in := utils.ByteToBits(h[:16])
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId},
		Input: in,
	}

	retrievePIRComplex(t, xof, db, q, match, "TestPIRComplex")
}

func TestPIRPointOneMb(t *testing.T) {
	dbLen := oneMB
	blockLen := testBlockLength * field.Bytes
	elemBitSize := 8
	numBlocks := dbLen / (elemBitSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	// functions defined in vpir_test.go
	xofDB := utils.RandomPRG()
	xof := utils.RandomPRG()

	db := database.CreateRandomBytes(xofDB, dbLen, nRows, blockLen)

	retrievePIRPoint(t, xof, db, numBlocks, "PIRPointOneMb")
}

func retrievePIRComplex(t *testing.T, rnd io.Reader, db *database.DB, q *query.ClientFSS, match interface{}, testName string) {
	c := client.NewPIRfss(rnd, &db.Info)
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
	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())

	// verify output
	count := 0
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
	require.Equal(t, count, res)

}

func retrievePIRPoint(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
	c := client.NewPIR(rnd, &db.Info)
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		in := make([]byte, 4)
		binary.BigEndian.PutUint32(in, uint32(i))
		queries, err := c.QueryBytes(in, 2)
		require.NoError(t, err)

		a0, err := s0.AnswerBytes(queries[0])
		require.NoError(t, err)
		a1, err := s1.AnswerBytes(queries[1])
		require.NoError(t, err)

		answers := [][]byte{a0, a1}

		res, err := c.ReconstructBytes(answers)
		//fmt.Println(res)
		require.NoError(t, err)
		require.Equal(t, db.Entries[i*db.BlockSize:(i+1)*db.BlockSize], res)
	}
	fmt.Printf("TotalCPU time %s: %.2fms\n", testName, totalTimer.Record())
}
