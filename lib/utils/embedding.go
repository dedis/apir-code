package utils

import (
	"encoding/binary"

	"github.com/si-co/vpir-code/lib/field"
)

// TODO: finish this
func EmbedKey(key []byte) []uint32 {
	out := make([]uint32, 0)
	bits := len(key) * 8
	remaining := bits
	for remaining > 0 {
		current := binary.BigEndian.Uint32(key[len(key)-field.Bytes:]) & 0x3fffffff
		out = append([]uint32{current}, out...)
		//key = key >> 30
		remaining -= 30
	}

	return out
}
