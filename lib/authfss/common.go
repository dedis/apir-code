package authfss

// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/common.go

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/field"
)

type Fss struct {
	// store keys used in fixedBlocks so that they can be sent to the server
	PrfKeys     [][]byte
	FixedBlocks []cipher.Block
	N           uint
	NumBits     uint   // number of bits in domain
	Temp        []byte // temporary slices so that we only need to allocate memory at the beginning
	Out         []byte

	BlockLength     int    // block length in number of elements
	OutConvertBlock []byte // to gather random bytes in convertBlock, allocate once for performance
}

//const initPRFLen uint = 4

// Structs for keys
type FssKeyEq2P struct {
	SInit   []byte
	TInit   byte
	CW      [][]byte // there are n
	FinalCW []uint32
}

type CWLt struct {
	cs [][]byte
	ct []uint8
	cv [][]uint32 // NOTE: elements of the group, i.e. F^(1+b)
}

type ServerKeyLt struct {
	s  [][]byte
	t  []uint8
	v  [][]uint32 // NOTE: elements of the group, i.e. F^(1+b)
	cw [][]CWLt   // Should be length n
}

// Helper functions

// 0th position is the most significant bit
// True if bit is 1 and False if bit is 0
// N is the number of bits in uint
func getBit(n uint64, pos, N uint) byte {
	return byte((n & (1 << (N - pos))) >> (N - pos))
}

// fixed key PRF (Matyas–Meyer–Oseas one way compression function)
// numBlocks represents the number
func prf(x []byte, aesBlocks []cipher.Block, numBlocks uint, temp, out []byte) {
	for i := uint(0); i < numBlocks; i++ {
		// get AES_k[i](x)
		aesBlocks[i].Encrypt(temp, x)
		// get AES_k[i](x) ^ x
		fastxor.Bytes(out[i*aes.BlockSize:], temp, x)
	}
}

func convertBlock(f Fss, x []byte, out []uint32) {
	// we can generate four uint32 numbers with a 16-bytes AES block
	prf(x, f.FixedBlocks, uint(len(out)/4), f.Temp, f.OutConvertBlock)
	field.ByteSliceToFieldElementSlice(out, f.OutConvertBlock)
}
