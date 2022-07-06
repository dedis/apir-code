package utils

import (
	"math"
	"math/rand"
	"time"
)

// Given a vector index returns the indices for the matrix representation
// of the vector
func VectorToMatrixIndices(i, numColumns int) (int, int) {
	return i / numColumns, i % numColumns
}

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

// source: https://stackoverflow.com/questions/43495745/how-to-generate-random-date-in-go-lang/43497333
// this is probably biased, but we don't care since it is only for tests
func Randate() time.Time {
	min := time.Date(2000, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2021, 12, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func Ranstring(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
