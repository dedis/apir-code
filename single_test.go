package main

// Test suite for the single-server VPIR scheme

import (
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	"github.com/tuneinsight/lattigo/v3/bfv"
)

func TestDHMatrixOneMb(t *testing.T) {
	dbLen := 1024 * 1024 // dbLen is specified in bits
	dbPRG := utils.RandomPRG()
	ecg := group.P256
	db := database.CreateRandomEllipticWithDigest(dbPRG, dbLen, ecg, true)
	fmt.Println("DB created")
	prg := utils.RandomPRG()
	retrieveBlocksDH(t, prg, db, "SingleMatrixOneMB")
}

func retrieveBlocksDH(t *testing.T, rnd io.Reader, db *database.Elliptic, testName string) {
	c := client.NewDH(rnd, &db.Info)
	s := server.NewDH(db)

	var i int
	totalTimer := monitor.NewMonitor()
	for j := 0; j < 10; j++ {
		i = rand.Intn(db.NumRows * db.NumColumns)
		query, err := c.QueryBytes(i)
		require.NoError(t, err)

		a, err := s.AnswerBytes(query)
		require.NoError(t, err)

		res, err := c.ReconstructBytes(a)
		require.NoError(t, err)
		require.Equal(t, db.Entries[i], res)
	}
	fmt.Printf("\nTotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}

func TestLatticeMatrixOneMb(t *testing.T) {
	//dbLen := 83886080 // specified in bits
	dbPRG := utils.RandomPRG()
	lens := []int{1000000, 10000000}

	for _, dbLen := range lens {
		db := database.CreateRandomRingDB(dbPRG, dbLen, true)
		fmt.Printf("DB of size %d created\n", dbLen)
		retrieveBlocksLattice(t, db, "LatticeMatrixOneMB")
	}
}

func retrieveBlocksLattice(t *testing.T, db *database.Ring, testName string) {
	c := client.NewLattice(&db.Info)
	s := server.NewLattice(db)

	encoder := bfv.NewEncoder(db.LatParams)
	totalTimer := monitor.NewMonitor()
	var i int
	for j := 0; j < 1; j++ {
		i = rand.Intn(db.NumRows * db.NumColumns)
		query, err := c.QueryBytes(i)
		require.NoError(t, err)

		a, err := s.AnswerBytes(query)
		require.NoError(t, err)

		res, err := c.ReconstructBytes(a)
		require.NoError(t, err)
		require.Equal(t, encoder.DecodeUintNew(db.Entries[i]), res)
	}

	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}
