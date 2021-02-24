package database

import (
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

func TestMerkle(t *testing.T) {
	rng := utils.RandomPRG()
	CreateRandomMultiBitMerkle(rng, 100000, 2, 10)
}
