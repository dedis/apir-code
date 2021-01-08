package client

import (
	"fmt"
	"testing"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

func TestVectorQuery(t *testing.T) {
	xof, err := blake2b.NewXOF(0, []byte("key"))
	require.NoError(t, err)

	numServers := 2
	rebalanced := false

	c := NewITMulti(xof, rebalanced)

	index := 0
	queries := c.Query(index, constants.BlockLength, numServers)

	// two servers
	require.Equal(t, 2, len(queries))

	// dbLenght
	require.Equal(t, constants.DBLength, len(queries[0]))

	// blockLength
	require.Equal(t, constants.BlockLength+1, len(queries[0][0]))

	// test secret sharing
	// only works with two servers!
	query := make([][]field.Element, constants.DBLength)
	for n := 0; n < constants.DBLength; n++ {
		query[n] = make([]field.Element, constants.BlockLength+1)
		for b := 0; b < constants.BlockLength+1; b++ {
			query[n][b].Add(&queries[0][n][b], &queries[1][n][b])

		}
	}

	fmt.Println("query:", query)
	fmt.Println("length:", len(query[0]))

	// first element of index = 0 should be field.One()
	require.Equal(t, field.One(), query[index][0])

	// index = 0 is the alpha polynomial, rest should be the zero vector
	for i := index + 1; i < constants.DBLength; i++ {
		for b := 0; b < constants.BlockLength; b++ {
			require.Equal(t, field.Zero(), query[i][b], "for i:", i, "and b:", b)
		}
	}

}
