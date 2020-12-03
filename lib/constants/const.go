package constants

import (
	"math"
	"math/big"
)

var BigZero = big.NewInt(0)
var BigOne = big.NewInt(1)

// scheme parameters
var Modulo = big.NewInt(int64(math.Pow(2, 64)) - 1)

const DBLength = 50000
