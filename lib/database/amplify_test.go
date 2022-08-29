package database

import (
	"fmt"
	"testing"
)

func TestDigestAmplify(t *testing.T) {
	tECC := 10
	data := []uint32{1, 2, 3, 4, 5}
	d := DigestAmplify(tECC, data)
	fmt.Println(d)
}
