package field

import (
	"fmt"
	"testing"
)

func TestMul(t *testing.T) {
	z := new(Element).SetOne()
	a := new(Element).SetOne()
	//b := new(Element).SetOne()
	z.Mul(a, a)
	fmt.Println(z)
}
