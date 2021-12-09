package server

import (
	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/database"
)

// Server is a scheme-agnostic VPIR server interface, implemented by both IT
// and DPF-based schemes
type Server interface {
	AnswerBytes([]byte) ([]byte, error)
	DBInfo() *database.Info
}

/*
%%	PIR primitives
*/
func answerPIR(q []byte, db *database.Bytes, NGoRoutines int) []byte {
	// we only use matrix db
	var prevPos, nextPos int
	out := make([]byte, db.NumRows*db.BlockSize)

	for i := 0; i < db.NumRows; i++ {
		for j := 0; j < db.NumColumns; j++ {
			nextPos += db.BlockLengths[i*db.NumColumns+j]
		}
		xorValues(db.Entries[prevPos:nextPos], db.BlockLengths[i*db.NumColumns:(i+1)*db.NumColumns], q, db.BlockSize, out[i*db.BlockSize:(i+1)*db.BlockSize])
		prevPos = nextPos
	}
	return out
}

// XORs entries and q block by block of size bl
func xorValues(entries []byte, blockLens []int, q []byte, bl int, out []byte) {
	pos := 0
	for j := range blockLens {
		if (q[j/8]>>(j%8))&1 == byte(1) {
			fastxor.Bytes(out, out, entries[pos:pos+blockLens[j]])
		}
		pos += blockLens[j]
	}
}
