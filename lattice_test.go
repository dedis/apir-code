package main

// Test suite for the single-server data retrieval, i.e. the classical PIR part
// of the single-server VPIR scheme.

import (
	"fmt"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestLatticeMatrixOneMb(t *testing.T) {
	dbLen := 838860800 // specified in bits
	dbPRG := utils.RandomPRG()
	db, _ := database.CreateRandomRingDB(dbPRG, dbLen, true)

	retrieveBlocksLattice(t, db, "LatticeMatrixOneMB")
}

func retrieveBlocksLattice(t *testing.T, db *database.Ring, testName string) {
	c := client.NewLattice(&db.Info)
	s := server.NewLattice(db)

	//encoder := bfv.NewEncoder(db.LatParams)
	totalTimer := monitor.NewMonitor()
	//for i := 0; i < db.NumRows*db.NumColumns; i++ {
	for i := 0; i < 10; i++ {
		query, err := c.QueryBytes(i)
		require.NoError(t, err)

		a, err := s.AnswerBytes(query)
		require.NoError(t, err)

		_, err = c.ReconstructBytes(a)
		require.NoError(t, err)
		//require.Equal(t, encoder.DecodeUintNew(db.Entries[i]), res)
	}

	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}

func TestLatticeWithDHTag(t *testing.T) {
	dbLen := 838860800 // specified in bits
	dbPRG := utils.RandomPRG()
	dbRing, data := database.CreateRandomRingDB(dbPRG, dbLen, true)
	dbElliptic := database.CreateEllipticWithDigestFromData(data, &dbRing.Info)

	prg := utils.RandomPRG()
	cL := client.NewLattice(&dbRing.Info)
	sL := server.NewLattice(dbRing)
	cE := client.NewDH(prg, &dbElliptic.Info)
	sE := server.NewDH(dbElliptic)

	//encoder := bfv.NewEncoder(dbRing.LatParams)
	totalTimer := monitor.NewMonitor()
	//for i := 0; i < dbRing.NumRows*dbRing.NumColumns; i++ {
	for i := 0; i < 1; i++ {
		fmt.Printf("%d ", i)

		queryL, err := cL.QueryBytes(i)
		require.NoError(t, err)
		queryE, err := cE.QueryBytes(i)
		require.NoError(t, err)

		aL, err := sL.AnswerBytes(queryL)
		require.NoError(t, err)
		aE, err := sE.AnswerBytes(queryE)
		require.NoError(t, err)

		column, err := cL.ReconstructBytes(aL)
		require.NoError(t, err)
		ctxSize := int(dbRing.LatParams.N()) * 2
		index := i % dbRing.NumColumns
		for j, col := range column {
			require.Equal(t, data[(j*dbRing.NumColumns+index)*ctxSize:(j*dbRing.NumColumns+index+1)*ctxSize], col)
		}
		_, err = cE.ReconstructBytes(aE, column)
		require.NoError(t, err)
	}

	fmt.Printf("\nTotalCPU time %s: %.1fms\n", "TestLatticeWithDHTag", totalTimer.Record())
}
