package constants

import (
	"math"
	"math/big"

	"github.com/si-co/vpir-code/lib/field"
)

const (
	DBLength    = 40000
	FieldSize   = 64
	BlockLength = 16
	BlockSizeSingleBit = 0
)

var (
	BigZero = big.NewInt(0)
	BigOne  = big.NewInt(1)

	// scheme parameters
	Modulo = big.NewInt(int64(math.Pow(2, FieldSize)) - 1)
	//Modulo = big.NewInt(math.MaxInt64)

	Zero field.Element
	One  field.Element
)

func init() {
	Zero.SetZero()
	One.SetOne()
}
