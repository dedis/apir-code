package dpf

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"

	"github.com/si-co/vpir-code/lib/field"
)

func TestAll(t *testing.T) {
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	require.NoError(t, err)
	rand, err := new(field.Element).SetRandom(xof)
	require.NoError(t, err)

	// Generate fss Keys on client
	fClient := ClientInitialize(6)
	// Test with if x = 10, evaluate to 2
	fssKeys := fClient.GenerateTreePF(10, rand)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	// Test 2-party Equality Function
	var ans0, ans1 *field.Element
	ans0 = fServer.EvaluatePF(0, fssKeys[0], 10)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 10)
	require.Equal(t, new(field.Element).Add(ans0, ans1).String(), rand.String())

	ans0 = fServer.EvaluatePF(0, fssKeys[0], 11)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 11)
	zero := new(field.Element).SetZero()
	require.Equal(t,
		new(field.Element).Add(ans0, ans1).String(),
		zero.String())

	ans0 = fServer.EvaluatePF(0, fssKeys[0], 9)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 9)
	require.Equal(t,
		new(field.Element).Add(ans0, ans1).String(),
		zero.String())
}

func TestVector(t *testing.T) {
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	require.NoError(t, err)
	rand, err := new(field.Element).SetRandom(xof)
	require.NoError(t, err)

	alpha := uint(129)
	nBits := uint(20)
	length := 10

	fClient := ClientInitialize(nBits)
	fssKeysVector := fClient.GenerateTreePFVector(alpha, rand, length)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	// compute expected vector
	expectedVector := make([]field.Element, length+1)
	expectedVector[0] = field.One()
	expectedVector[1] = *rand
	for i := 2; i < len(expectedVector); i++ {
		expectedVector[i].Mul(&expectedVector[i-1], rand)
	}

	zero := field.Zero()
	for i := uint(0); i < (1 << nBits); i++ {
		// Test 2-party vector Equality Function

		// generate vector of answers
		ans0 := fServer.EvaluatePFVector(0, fssKeysVector[0], i)
		ans1 := fServer.EvaluatePFVector(1, fssKeysVector[1], i)

		// test all elements of vector
		for j := range ans0 {
			val := new(field.Element).Add(ans0[j], ans1[j]).String()
			if i == alpha {
				require.Equal(t, expectedVector[j].String(), val)
			} else {
				require.Equal(t, zero.String(), val)
			}
		}
	}

}

func TestEval(t *testing.T) {
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	require.NoError(t, err)
	rand, err := new(field.Element).SetRandom(xof)
	require.NoError(t, err)

	alpha := uint(129)
	nBits := uint(20)

	fClient := ClientInitialize(nBits)
	fssKeys := fClient.GenerateTreePF(alpha, rand)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	zero := field.Zero()
	for i := uint(0); i < (1 << nBits); i++ {
		// Test 2-party Equality Function
		var ans0, ans1 *field.Element
		ans0 = fServer.EvaluatePF(0, fssKeys[0], i)
		ans1 = fServer.EvaluatePF(1, fssKeys[1], i)

		if i == alpha {
			require.Equal(t, new(field.Element).Add(ans0, ans1).String(), rand.String())
		} else {
			require.Equal(t, new(field.Element).Add(ans0, ans1).String(), zero.String())
		}
	}
}
