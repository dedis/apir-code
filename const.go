package main

import (
	"math"
	"math/big"
)

var bigZero = big.NewInt(0)
var bigOne = big.NewInt(1)

var MODULO = big.NewInt(int64(math.Pow(2, 64)) - 1)
var DBLength = 50000
var servers = 3
