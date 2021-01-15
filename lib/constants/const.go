package constants

import (
	"github.com/si-co/vpir-code/lib/field"
)

const (
	DBLength             = 40000
	BlockLength          = 16
	ChunkBytesLength     = 15
	SingleBitBlockLength = 0
)

var (
	Zero field.Element
	One  field.Element
)

func init() {
	Zero.SetZero()
	One.SetOne()
}
