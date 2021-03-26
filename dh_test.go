package main

import (
	"fmt"
	"io"
	"testing"

	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestDHMatrixOneMb(t *testing.T) {
	dbLen := 8000000
	dbPRG := utils.RandomPRG()
	blockLen := 16
	ecg := group.P256
	db := database.CreateRandomEllipticWithDigest(dbPRG, ecg, dbLen, blockLen, true)

	prg := utils.RandomPRG()

	retrieveBlocksDH(t, prg, db, "SingleMatrixOneMB")
}

func retrieveBlocksDH(t *testing.T, rnd io.Reader, db *database.Elliptic, testName string) {
	c := client.NewDH(rnd, &db.Info)
	s := server.NewDH(db)

	totalTimer := monitor.NewMonitor()
	//for i := 0; i < db.NumRows*db.NumColumns; i++ {
	for i := 0; i < 10; i++ {
		fmt.Printf("%d ", i)
		query, err := c.QueryBytes(i)
		require.NoError(t, err)

		_, err = s.AnswerBytes(query)
		require.NoError(t, err)

		//_, err = c.ReconstructBytes(a, nil)
		//require.NoError(t, err)
	}

	fmt.Printf("\nTotalCPU time %s: %.1fms\n", testName, totalTimer.Record())
}
