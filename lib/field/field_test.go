package field

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandVectorWithPRG(t *testing.T) {
	vec := RandVectorWithPRG(100000, rand.Reader)
	for _, v := range vec {
		require.Less(t, v, ModP)
	}
}
