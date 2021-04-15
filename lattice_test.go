package main

// Test suite for the single-server data retrieval, i.e. the classical PIR part
// of the single-server VPIR scheme.

import (
	"fmt"
	"github.com/ldsec/lattigo/v2/bfv"
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestLatticeMatrixOneMb(t *testing.T) {
	//dbLen := 83886080 // specified in bits
	dbPRG := utils.RandomPRG()
	lens := []int{1000000, 10000000}

	for _, dbLen := range lens {
		db := database.CreateRandomRingDB(dbPRG, dbLen, true)
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
