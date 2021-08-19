package field

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"testing"
)


func TestRandVectorWithPRG(t *testing.T) {
	vec := RandVectorWithPRG(100000, rand.Reader)
	for _, v := range vec {
		require.Less(t, v, ModP)
	}
}
