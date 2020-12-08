package constants

import (
	"math"
	"math/big"
)

const (
	DBLength = 50000
	FieldSize = 64
)

var (
	BigZero = big.NewInt(0)
	BigOne = big.NewInt(1)

	// scheme parameters
	Modulo = big.NewInt(int64(math.Pow(2, FieldSize)) - 1)
)


