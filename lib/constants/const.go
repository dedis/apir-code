package constants

import (
	"github.com/si-co/vpir-code/lib/field"
)

const (
	//BlockLength          = 16
	ChunkBytesLength     = 15
	SingleBitBlockLength = 0
	IDLengthBytes        = 45
)

var (
	Zero field.Element
	One  field.Element
)

func init() {
	Zero.SetZero()
	One.SetOne()
}
