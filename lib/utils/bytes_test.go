package utils

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytePadding(t *testing.T) {
	b := make([]byte, 100)
	rand.Read(b)
	//b := []byte{255, 255}
	//fmt.Println("original", b)

	b1 := BitsPadded(b)
	//fmt.Println("bits padded:", b1)

	packedBits := PackBits(b1)
	//fmt.Println("packedBits:", packedBits)

	bitsRecovered := BytesTobits(packedBits)

	b2 := BytesUnpadded(bitsRecovered)
	//fmt.Println("recovered:", b2)

	require.Equal(t, b, b2)
}
