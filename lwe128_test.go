package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestLWEMatrixOneMb128(t *testing.T) {
	dbLen := 1024 * 1024 // dbLen is specified in bits
	db := database.CreateRandomBinaryLWEWithLength128(utils.RandomPRG(), dbLen)
	p := utils.ParamsWithDatabaseSize128(db.Info.NumRows, db.Info.NumColumns)
	retrieveBlocksLWE128(t, db, p, "SingleMatrixLWEOneMb_128")
}

func TestDefaultLWE128(t *testing.T) {
	// get default parameters
	p := utils.ParamsDefault128()
	db := database.CreateRandomBinaryLWE128(utils.RandomPRG(), p.L, p.M)
	retrieveBlocksLWE128(t, db, p, "SingleDefaultLWE128")
}

func retrieveBlocksLWE128(t *testing.T, db *database.LWE128, params *utils.ParamsLWE, testName string) {
	c := client.NewLWE128(utils.RandomPRG(), &db.Info, params)
	s := server.NewLWE128(db)

	totalTimer := monitor.NewMonitor()
	for j := 0; j < 10; j++ {
		i := rand.Intn(params.L * params.M)
		query, err := c.QueryBytes(i)
		require.NoError(t, err)

		a, err := s.AnswerBytes(query)
		require.NoError(t, err)

		res, err := c.ReconstructBytes(a)
		require.NoError(t, err)
		require.Equal(t, uint32(db.Matrix.Get(utils.VectorToMatrixIndices(i, db.Info.NumColumns)).Hi), res)
	}
	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())

}
