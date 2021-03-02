package merkle

import (
	"math/rand"
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeProof(t *testing.T) {
	rng := utils.RandomPRG()
	data := make([][]byte, rand.Intn(501))
	for i := range data {
		d := make([]byte, 32)
		rng.Read(d)
		data[i] = d
	}

	// create the tree
	tree, err := New(data)
	require.NoError(t, err)

	// generate a proof for random element
	proof, err := tree.GenerateProof(data[rand.Intn(len(data))])
	require.NoError(t, err)

	// encode the proof
	b := EncodeProof(proof)

	// decode proof
	p := DecodeProof(b)

	require.Equal(t, *proof, *p)
}
