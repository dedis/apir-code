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

func TestLWE(t *testing.T) {
	dbLen := 1024 * 1024 // dbLen is specified in bits
	db := database.CreateRandomBinaryLWEWithLength(utils.RandomPRG(), dbLen)
	p := utils.ParamsWithDatabaseSize(db.Info.NumRows, db.Info.NumColumns)
	retrieveBlocksLWE(t, db, p, "TestLWE")
}

func retrieveBlocksLWE(t *testing.T, db *database.LWE, params *utils.ParamsLWE, testName string) {
	c := client.NewLWE(utils.RandomPRG(), &db.Info, params)
	s := server.NewLWE(db)

	totalTimer := monitor.NewMonitor()
	repetitions := 100
	for j := 0; j < repetitions; j++ {
		i := rand.Intn(params.L * params.M)
		query, err := c.QueryBytes(i)
		require.NoError(t, err)

		a, err := s.AnswerBytes(query)
		require.NoError(t, err)

		res, err := c.ReconstructBytes(a)
		require.NoError(t, err)
		require.Equal(t, uint32(db.Matrix.Get(utils.VectorToMatrixIndices(i, db.Info.NumColumns))), res)
	}
	fmt.Printf("Total CPU time %s: %.1fms\n", testName, totalTimer.Record())
}
