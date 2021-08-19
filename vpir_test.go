package main

// Test suite for integrated VPIR.

import (
	"io"
	"testing"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

const (
	oneB            = 8
	oneKB           = 1024 * oneB
	oneMB           = 1024 * oneKB
	testBlockLength = 64
)

func TestMultiBitVPIR(t *testing.T) {
	numIdentifiers := 1000
	rndDB := utils.RandomPRG()
	_, err := database.CreateRandomDB(rndDB, numIdentifiers)
	require.NoError(t, err)
}

func getXof(t *testing.T, key string) io.Reader {
	return utils.RandomPRG()
}
