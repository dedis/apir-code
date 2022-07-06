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

func TestLWEMatrixOneMb(t *testing.T) {
	p := utils.ParamsWithDatabaseSize(1024, 1024)
	db := database.CreateRandomBinaryLWE(utils.RandomPRG(), p.L, p.M)
	retrieveBlocksLWE(t, db, p, "SingleMatrixLWEOneMb")
}

func TestDefaultLWE(t *testing.T) {
	// get default parameters
	p := utils.ParamsDefault()
	db := database.CreateRandomBinaryLWE(utils.RandomPRG(), p.L, p.M)
	retrieveBlocksLWE(t, db, p, "SingleDefaultLWE")
}

func retrieveBlocksLWE(t *testing.T, db *database.LWE, params *utils.ParamsLWE, testName string) {
	c := client.NewLWE(utils.RandomPRG(), &db.Info, params)
	s := server.NewLWE(db)

	totalTimer := monitor.NewMonitor()
	for j := 0; j < 10; j++ {
		i := rand.Intn(params.L * params.M)
		query, err := c.QueryBytes(i)
		require.NoError(t, err)

		a, err := s.AnswerBytes(query)
		require.NoError(t, err)

		res, err := c.ReconstructBytes(a)
		require.NoError(t, err)
		require.Equal(t, db.Matrix.Get(utils.VectorToMatrixIndices(i, db.Info.NumColumns)), res)
	}
	fmt.Printf("\nTotalCPU time %s: %.1fms\n", testName, totalTimer.Record())

}
