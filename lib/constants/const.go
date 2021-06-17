package constants

import (
	"github.com/si-co/vpir-code/lib/field"
)

const (
	// number of bytes embedded in each field element.
	//ChunkBytesLength     = 15
	//ChunkBytesLength     = 16
	ChunkBytesLength     = 4
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
