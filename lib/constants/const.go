package constants

import (
	"github.com/ncw/gmp"
	"math"
)

const (
	DBLength = 50000
	FieldSize = 64
)

var (
	BigZero = gmp.NewInt(0)
	BigOne = gmp.NewInt(1)

	// scheme parameters
	Modulo = gmp.NewInt(int64(math.Pow(2, FieldSize)) - 1)
)


