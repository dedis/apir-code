package utils

import "math"

// MaxBytesLength get maximal []byte length in map[int][]byte
func MaxBytesLength(in map[int][]byte) int {
	max := 0
	for _, v := range in {
		if len(v) > max {
			max = len(v)
		}
	}

	return max
}

// Divides dividend by divisor and rounds up the result
func DivideAndRoundUp(dividend, divisor int) int {
	return int(math.Ceil(float64(dividend) / float64(divisor)))
}

// Increase num to the next perfect square.
// If the square root is a whole number, do not modify anything.
// Otherwise, return the square of the square root + 1.
func IncreaseToNextSquare(num *int) {
	i, f := math.Modf(math.Sqrt(float64(*num)))
	if f == 0 {
		return
	}
	*num = int(math.Pow(i+1, 2))
}
