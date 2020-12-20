package field

import (
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// source: https://tools.ietf.org/html/rfc8452#section-7
func TestNewElement(t *testing.T) {
	// s = 11111111000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010010000
	x, err := hex.DecodeString("ff000000000000000000000000000090")
	require.NoError(t, err)
	e := NewElement(x)
	require.Equal(t, strconv.FormatUint(e.element.low, 2), "1111111100000000000000000000000000000000000000000000000000000000")
	require.Equal(t, strconv.FormatUint(e.element.high, 2), "10010000")
	require.Equal(t, uint64(0x0), (e.element.high >> 63) & 1)
	require.Equal(t, uint64(0x1), (e.element.high >> 7) & 1)
	require.Equal(t, uint64(0x0), (e.element.high >> 8) & 1)
	require.Equal(t, uint64(0x0), e.element.low & 1)
	require.Equal(t, uint64(0x1), e.element.low >> 63)
	require.Equal(t, uint64(0x1), (e.element.low >> 56) & 1)
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
	x, err := hex.DecodeString("20000000000000000000000000000000")
	require.NoError(t, err)
	y, err := hex.DecodeString("10000000000000000000000000000000")
	require.NoError(t, err)

	res := &Element{}
	res.Mul(NewElement(x), NewElement(y))

	require.Equal(t, "20000000000000000000000000000000", res.HexString())
}
