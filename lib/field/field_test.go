package field

import (
  "bytes"
	"testing"

	"github.com/stretchr/testify/require"
)



func TestSquare(t *testing.T) {
  g := Random()
  g2 := Mul(g, g)
  g3 := Mul(g, g2)
  g4 := Mul(g2, g2)
  g4_b := Mul(g3, g)

	require.Equal(t, g4, g4_b)
}

func TestRandom(t *testing.T) {
  var zeros [NumBytes]byte
  for i := 0; i < 100; i++ {
    b := Random().Bytes()
    require.Equal(t, NumBytes, len(b))
    require.Equal(t, false, bytes.Compare(b[:], zeros[:]) == 0)
  }
}

func TestNegate(t *testing.T) {
  for i := 0; i < 100; i++ {
    r := Random()
    rn := Add(r, Zero())
    rn.Negate()

    sum := Add(r, rn)
    require.Equal(t, Zero(), sum)
  }
}

func TestMul(t *testing.T) {
  for i := 0; i < 100; i++ {
    a := Random()
    b := Random()
    c := Random()

    b_plus_c := Add(b, c)
    a_times_b := Mul(a, b)
    a_times_c := Mul(a, c)

    abc := Mul(a, b_plus_c)
    abc_prime := Add(a_times_b, a_times_c)

    require.Equal(t, abc.Bytes(), abc_prime.Bytes())
  }
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

