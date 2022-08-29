package main

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

func TestAmplifyOneMb(t *testing.T) {
	threshold := 8
	dbLen := 1024 * 1024 // dbLen is specified in bits
	db := database.CreateRandomBinaryLWEWithLength(utils.RandomPRG(), dbLen)
	p := utils.ParamsWithDatabaseSize(db.Info.NumRows, db.Info.NumColumns)

	retrieveBlocksAmplify(t, db, p, threshold, "TestAmplifyOneMb")
}

func retrieveBlocksAmplify(t *testing.T, db *database.LWE, params *utils.ParamsLWE, threshold int, testName string) {
	c := client.NewAmplify(utils.RandomPRG(), &db.Info, params, threshold)
	s := server.NewAmplify(db)

	totalTimer := monitor.NewMonitor()
	repetitions := 100
	for k := 0; k < repetitions; k++ {
		i := 112
		j := 14
		query := c.Query(i, j)
		//require.NoError(t, err)

		a := s.Answer(query)
		//require.NoError(t, err)

		res, err := c.Reconstruct(a)
		require.NoError(t, err)
		require.Equal(t, db.Matrix.Get(i, j), res)
	}
	fmt.Printf("TotalCPU time %s: %.1fms\n", testName, totalTimer.Record())

}
