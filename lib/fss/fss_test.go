package libfss

// Source: https://github.com/frankw2/libfss/blob/master/go/test_fss/test_fss.go

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFss(t *testing.T) {
	// Generate fss Keys on client
	fClient := ClientInitialize(6)
	// Test with if x = 10, evaluate to 2
	fssKeys := fClient.GenerateTreePF(10, 2)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	// Test 2-party Equality Function
	var ans0, ans1 int = 0, 0
	ans0 = fServer.EvaluatePF(0, fssKeys[0], 10)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 10)
	require.NotEqual(t, 0, ans0+ans1)

	ans0 = fServer.EvaluatePF(0, fssKeys[0], 11)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 11)
	require.Equal(t, 0, ans0+ans1)

	ans0 = fServer.EvaluatePF(0, fssKeys[0], 9)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 9)
	require.Equal(t, 0, ans0+ans1)

	// Test 2-party Less than Function
	// Test if x < 10, evaluate to 2
	fssKeysLt := fClient.GenerateTreeLt(10, 2)

	var anslt0, anslt1 uint = 0, 0
	anslt0 = fServer.EvaluateLt(fssKeysLt[0], 8)
	anslt1 = fServer.EvaluateLt(fssKeysLt[1], 8)
	require.NotEqual(t, uint(0), anslt0-anslt1)

	anslt0 = fServer.EvaluateLt(fssKeysLt[0], 11)
	anslt1 = fServer.EvaluateLt(fssKeysLt[1], 11)
	require.Equal(t, uint(0), anslt0-anslt1)
}
