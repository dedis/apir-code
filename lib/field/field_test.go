package field

import (
	"encoding/hex"
	"testing"
)

// source: https://tools.ietf.org/html/rfc8452#section-7
func TestAdd(t *testing.T) {
	x, err := hex.DecodeString("66e94bd4ef8a2c3b884cfa59ca342b2e")
	if err != nil {
		panic(err)
	}
	y, err := hex.DecodeString("ff000000000000000000000000000000")
	if err != nil {
		panic(err)
	}

	res := &FieldElement{}

	res.Add(NewByte(x), NewByte(y))

	if hex.EncodeToString(res.Bytes()) != "99e94bd4ef8a2c3b884cfa59ca342b2e" {
		t.Fatal("wrong result for Add")
	}
}

func TestMul(t *testing.T) {
	x, err := hex.DecodeString("66e94bd4ef8a2c3b884cfa59ca342b2e")
	if err != nil {
		panic(err)
	}
	y, err := hex.DecodeString("ff000000000000000000000000000000")
	if err != nil {
		panic(err)
	}

	res := &FieldElement{}

	res.Mul(NewByte(x), NewByte(y))

	if hex.EncodeToString(res.Bytes()) != "37856175e9dc9df26ebc6d6171aa0ae9" {
		t.Fatal("wrong result")
	}
}
