package dpf

import (
  "testing"

	"github.com/stretchr/testify/require"

	"github.com/si-co/vpir-code/lib/field"
)

func TestAll(t *testing.T) {
  rand := field.Random()

	// Generate fss Keys on client
	fClient := ClientInitialize(6)
	// Test with if x = 10, evaluate to 2
	fssKeys := fClient.GenerateTreePF(10, &rand)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

	// Test 2-party Equality Function
  var ans0, ans1 *field.Element
	ans0 = fServer.EvaluatePF(0, fssKeys[0], 10)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 10)
  require.Equal(t, field.Add(*ans0,*ans1).String(), rand.String())

	ans0 = fServer.EvaluatePF(0, fssKeys[0], 11)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 11)
  require.Equal(t, field.Add(*ans0,*ans1).String(), field.Zero().String())

	ans0 = fServer.EvaluatePF(0, fssKeys[0], 9)
	ans1 = fServer.EvaluatePF(1, fssKeys[1], 9)
  require.Equal(t, field.Add(*ans0,*ans1).String(), field.Zero().String())
}

func TestEval(t *testing.T) {
  rand := field.Random()
  alpha := uint(129)
  nBits := uint(20)

	fClient := ClientInitialize(nBits)
	fssKeys := fClient.GenerateTreePF(alpha, &rand)

	// Simulate server
	fServer := ServerInitialize(fClient.PrfKeys, fClient.NumBits)

  zero := field.Zero()
  for i := uint(0); i < (1 << nBits); i += 1 {
    // Test 2-party Equality Function
    var ans0, ans1 *field.Element
    ans0 = fServer.EvaluatePF(0, fssKeys[0], i)
    ans1 = fServer.EvaluatePF(1, fssKeys[1], i)

    if i == alpha {
      require.Equal(t, field.Add(*ans0,*ans1).String(), rand.String())
    } else {
      require.Equal(t, field.Add(*ans0,*ans1).String(), zero.String())
    }
  }
}
