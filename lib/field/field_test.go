package field

import (
  "bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)


func TestBytes(t *testing.T) {
  s := "1f000000000000000000000000000096"
	x, err := hex.DecodeString(s)
	require.NoError(t, err)
	e := NewElement(x)
	require.Equal(t, s, e.HexString())
}

func TestAdd(t *testing.T) {
	x, err := hex.DecodeString("66e94bd4ef8a2c3b884cfa59ca342b2e")
	require.NoError(t, err)
	y, err := hex.DecodeString("ff000000000000000000000000000000")
	require.NoError(t, err)

	res := gcmAdd(NewElement(x), NewElement(y))

	require.Equal(t, "99e94bd4ef8a2c3b884cfa59ca342b2e", res.HexString())
}

func TestSquare(t *testing.T) {
  g := Gen()
  g2 := Mul(g, g)
  g3 := Mul(g, g2)
  g4 := Mul(g2, g2)
  g4_b := Mul(g3, g)

	require.Equal(t, g4, g4_b)
}

func TestRandom(t *testing.T) {
  var zeros [16]byte
  for i := 0; i < 100; i++ {
    b := Random().Bytes()
    require.Equal(t, 16, len(b))
    require.Equal(t, false, bytes.Compare(b[:], zeros[:]) == 0)
  }
}

func TestMul(t *testing.T) {
  a := Random()
  b := Random()
  c := Random()

  b_plus_c := gcmAdd(b, c)
  a_times_b := Mul(a, b)
  a_times_c := Mul(a, c)

  abc := Mul(a, b_plus_c)
  abc_prime := gcmAdd(a_times_b, a_times_c)

	require.Equal(t, abc.Bytes(), abc_prime.Bytes())
}

func TestMulCommute(t *testing.T) {
  r1 := Random()
  r2 := Random()

	res := Mul(r1, r2)
	res2 := Mul(r2, r1)

	require.Equal(t, res.Bytes(), res2.Bytes())
}

func TestMulZero(t *testing.T) {
	res := Mul(One(), Zero())
	require.Equal(t, Zero().Bytes(), res.Bytes())
}

func TestMulOneOne(t *testing.T) {
  left := One()
  right := One()
  res := Mul(left, right)
  res = Mul(left, res)
	require.Equal(t, One().Bytes(), res.Bytes())

}

func TestMulOne(t *testing.T) {
	x := One()
	y, err := hex.DecodeString("02400000000000000000000000000000")
	require.NoError(t, err)
	res := Mul(x, NewElement(y))
	require.Equal(t, "02400000000000000000000000000000", res.HexString())
}

func TestSimpleMul(t *testing.T) {
	x, err := hex.DecodeString("02400000000000000000000000000000")
	require.NoError(t, err)
	//y, err := hex.DecodeString("01000000000000000000000000000000")
	//require.NoError(t, err)

	res := Mul(NewElement(x), One())

	require.Equal(t, NewElement(x).Bytes(), res.Bytes())
	require.Equal(t, "02400000000000000000000000000000", res.HexString())
}
