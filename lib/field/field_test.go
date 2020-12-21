package field

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

// source: https://tools.ietf.org/html/rfc8452#section-7
func TestNewElement(t *testing.T) {
	// s = 00011111000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010010000
	x, err := hex.DecodeString("1f000000000000000000000000000090")
	require.NoError(t, err)
	e := NewElement(x)
	require.Equal(t, uint64(0x0), (e.value.high >> 63) & 1)
	require.Equal(t, uint64(0x1), (e.value.high >> 7) & 1)
	require.Equal(t, uint64(0x0), (e.value.high >> 8) & 1)
	require.Equal(t, uint64(0x0), e.value.low & 1)
	require.Equal(t, uint64(0x1), e.value.low >> 60)
	require.Equal(t, uint64(0x0), (e.value.low >> 55) & 1)
}

func TestAdd(t *testing.T) {
	x, err := hex.DecodeString("66e94bd4ef8a2c3b884cfa59ca342b2e")
	require.NoError(t, err)
	y, err := hex.DecodeString("ff000000000000000000000000000000")
	require.NoError(t, err)

	res := &Element{}
	res.Add(NewElement(x), NewElement(y))

	require.Equal(t, "99e94bd4ef8a2c3b884cfa59ca342b2e", res.HexString())
}

//func TestMul(t *testing.T) {
//	x, err := hex.DecodeString("66e94bd4ef8a2c3b884cfa59ca342b2e")
//	require.NoError(t, err)
//	y, err := hex.DecodeString("ff000000000000000000000000000000")
//	require.NoError(t, err)
//
//	res := &Element{}
//	res.Mul(NewElement(x), NewElement(y))
//
//	require.Equal(t, "37856175e9dc9df26ebc6d6171aa0ae9", res.HexString())
//}

func TestSimpleMul(t *testing.T) {
	x, err := hex.DecodeString("02000000000000000000000000000000")
	require.NoError(t, err)
	y, err := hex.DecodeString("01000000000000000000000000000000")
	require.NoError(t, err)

	res := &Element{}
	res.Mul(NewElement(x), NewElement(y))

	require.Equal(t, "02000000000000000000000000000000", res.HexString())
}
