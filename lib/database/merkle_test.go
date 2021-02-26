package database

import (
	"fmt"
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
	merkletree "github.com/wealdtech/go-merkletree"
)

func TestMerkle(t *testing.T) {
	rng := utils.RandomPRG()
	CreateRandomMultiBitMerkle(rng, 10000000, 30, 10)
}

func TestEncodeDecodeProof(t *testing.T) {
	rng := utils.RandomPRG()
	data := make([][]byte, 100)
	for i := range data {
		d := make([]byte, 32)
		rng.Read(d)
		data[i] = d
	}

	// create the tree
	tree, err := merkletree.New(data)
	require.NoError(t, err)

	// generate a proof for element 50
	proof, err := tree.GenerateProof(data[50])
	fmt.Println(proof)
	require.NoError(t, err)

	// encode the proof
	b := encodeProof(proof)

	// decode proof
	p := DecodeProof(b)

	require.Equal(t, *proof, *p)

}
