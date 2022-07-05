package main

import (
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestDefaultLWE(t *testing.T) {
	// get default parameters
	p := utils.ParamsDefault()

	// create random database with default parameters
	db := database.CreateRandomBinaryLWE(utils.RandomPRG(), p.L, p.M)

	// initialize client
	c := client.NewLWE(utils.RandomPRG(), &db.Info)

	// initialize server
	s := server.NewLWE(db)

	// query
	for k := 0; k < 100; k++ {
		i := rand.Intn(p.L * p.M)
		q, err := c.QueryBytes(i)
		require.NoError(t, err)

		a, err := s.AnswerBytes(q)
		require.NoError(t, err)

		res, err := c.ReconstructBytes(a)
		require.NoError(t, err)

		require.Equal(t, db.Matrix.Get(i/db.Info.NumColumns, i%db.Info.NumColumns), res)
	}
}
