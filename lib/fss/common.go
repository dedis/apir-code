package fss

// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/common.go

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"

	"github.com/lukechampine/fastxor"
)

var PrfKeys [][]byte

type Fss struct {
	FixedBlocks []cipher.Block
	N           uint
	NumBits     uint   // number of bits in domain
	Temp        []byte // temporary slices so that we only need to allocate memory at the beginning
	Out         []byte
}

const initPRFLen uint = 4

// Structs for keys

type FssKeyEq2P struct {
	SInit   []byte
	TInit   byte
	CW      [][]byte // there are n
	FinalCW int
}

type CWLt struct {
	cs [][]byte
	ct []uint8
	cv []uint
}

type ServerKeyLt struct {
	s  [][]byte
	t  []uint8
	v  []uint
	cw [][]CWLt // Should be length n
}

type FssKeyEqMP struct {
	NumParties uint
	CW         [][]uint32 //Assume CW is 32-bit because f.M is 4. If you change f.M, you should change this
	Sigma      [][]byte
}

func init() {
	PrfKeys = [][]byte{
		{2, 16, 223, 155, 240, 218, 18, 217, 66, 61, 95, 162, 213, 195, 169, 50},
		{130, 178, 43, 30, 226, 225, 106, 13, 196, 22, 96, 191, 75, 100, 87, 221},
		{227, 121, 10, 139, 215, 136, 201, 227, 253, 210, 170, 246, 215, 213, 65, 69},
		{49, 194, 90, 224, 41, 253, 48, 252, 55, 167, 51, 93, 246, 176, 38, 220}}
}

// Helper functions

func randomCryptoInt() uint {
	b := make([]byte, 8)
	rand.Read(b)
	ans, _ := binary.Uvarint(b)
	return uint(ans)
}

// fixed key PRF (Matyas–Meyer–Oseas one way compression function)
// numBlocks represents the number
func prf(x []byte, aesBlocks []cipher.Block, numBlocks uint, temp, out []byte) {
	// If request blocks greater than actual needed blocks, grow output array
	// if numBlocks > initPRFLen {
	// 	out = make([]byte, numBlocks*aes.BlockSize)
	// }
	for i := uint(0); i < numBlocks; i++ {
		// get AES_k[i](x)
		aesBlocks[i].Encrypt(temp, x)
		// get AES_k[i](x) ^ x
		fastxor.Bytes(out[i*aes.BlockSize:], temp, x)
	}
}
