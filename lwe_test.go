package main

import (
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/matrix"
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
	i := 2
	j := 7
	q := c.Query(i, j)
	qBytes := matrix.MatrixToBytes(q)

	a, err := s.AnswerBytes(qBytes)
	require.NoError(t, err)

	res, err := c.ReconstructBytes(a)
	require.NoError(t, err)

	require.Equal(t, db.Matrix.Get(i, j), res)
}
