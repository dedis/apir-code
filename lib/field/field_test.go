package field

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

// source: https://tools.ietf.org/html/rfc8452#section-7
func TestNewElement(t *testing.T) {
	// s = 11111000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001101001
	x, err := hex.DecodeString("1f000000000000000000000000000096")
	require.NoError(t, err)
	e := NewElement(x)
	//fmt.Println(strconv.FormatUint(e.value.low, 2))
	//fmt.Println(strconv.FormatUint(e.value.high, 2))
	require.Equal(t, uint64(0x1), (e.value.low>>63)&1)
	require.Equal(t, uint64(0x1), (e.value.low>>59)&1)
	require.Equal(t, uint64(0x0), (e.value.low>>58)&1)
	require.Equal(t, uint64(0x1), e.value.high&1)
	require.Equal(t, uint64(0x1), (e.value.high>>6)&1)
	require.Equal(t, uint64(0x0), (e.value.high>>7)&1)
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

func TestMul(t *testing.T) {
	x, err := hex.DecodeString("66e94bd4ef8a2c3b884cfa59ca342b2e")
	require.NoError(t, err)
	y, err := hex.DecodeString("ff000000000000000000000000000000")
	require.NoError(t, err)

	res := &Element{}
	res.Mul(NewElement(x), NewElement(y))

	expectedString := "37856175e9dc9df26ebc6d6171aa0ae9"
	expected, err := hex.DecodeString(expectedString)
	require.NoError(t, err)

	require.Equal(t, expected, res.Bytes())
	require.Equal(t, expectedString, res.HexString())
}

func TestSimpleMul(t *testing.T) {
	x, err := hex.DecodeString("02000000000000000000000000000000")
	require.NoError(t, err)
	y, err := hex.DecodeString("01000000000000000000000000000000")
	require.NoError(t, err)

	res := &Element{}
	res.Mul(NewElement(x), NewElement(y))

	require.Equal(t, "02000000000000000000000000000000", res.HexString())
}
