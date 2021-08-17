package client

import (
	"testing"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

const (
	oneB            = 8
	oneKB           = 1024 * oneB
	oneMB           = 1024 * oneKB
	testBlockLength = 16
)

func TestSecretSharing(t *testing.T) {
	dbLen := oneMB
	nRows := 1

	xofDB := utils.RandomPRG()
	xof := utils.RandomPRG()

	db, err := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, testBlockLength)
	require.NoError(t, err)

	c := NewIT(xof, &db.Info)

	// dumm query to generate client state
	_ = c.Query(0, 2)

	shares, err := c.secretShare(2)
	require.NoError(t, err)

	stillInBlock := 0
	for i := range shares[0] {
		result := (shares[0][i] + shares[1][i]) % field.ModP
		if stillInBlock < testBlockLength+1 {
			require.NotEqual(t, uint32(0), result)
		} else {
			require.Equal(t, uint32(0), result)
		}

		stillInBlock++
	}
}
