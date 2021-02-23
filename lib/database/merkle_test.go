package database

import (
	"testing"

	"github.com/si-co/vpir-code/lib/utils"
)

func TestMerkle(t *testing.T) {
	rng := utils.RandomPRG()
	CreateRandomMultiBitMerkle(rng, 10, 1, 10)
}
