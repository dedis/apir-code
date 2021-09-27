package fss

// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/common.go

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/field"
)

var PrfKeys [][]byte

type Fss struct {
	FixedBlocks []cipher.Block
	N           uint
	NumBits     uint   // number of bits in domain
	Temp        []byte // temporary slices so that we only need to allocate memory at the beginning
	Out         []byte

	BlockLength     int    // block length in number of elements
	OutConvertBlock []byte // to gather random bytes in convertBlock, allocate once for performance
}

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

func init() {
	PrfKeys = [][]byte{
		{2, 16, 223, 155, 240, 218, 18, 217, 66, 61, 95, 162, 213, 195, 169, 50},
		{130, 178, 43, 30, 226, 225, 106, 13, 196, 22, 96, 191, 75, 100, 87, 221},
		{227, 121, 10, 139, 215, 136, 201, 227, 253, 210, 170, 246, 215, 213, 65, 69},
		{49, 194, 90, 224, 41, 253, 48, 252, 55, 167, 51, 93, 246, 176, 38, 220}}
}

// Helper functions

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
	field.BytesToElements(out, f.OutConvertBlock)
}
