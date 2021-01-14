package dpf

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"

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
}

const initPRFLen uint = 4

// Structs for keys for DPF with output in F
type FssKeyEq2P struct {
	SInit   []byte
	TInit   byte
	CW      [][]byte // there are n
	FinalCW field.Element
}

// Structs for keys for DPF with output in F^(length)
type FssKeyVectorEq2P struct {
	SInit   []byte
	TInit   byte
	CW      [][]byte // there are n
	FinalCW []field.Element
}

// Helper functions

func randomCryptoInt() uint {
	b := make([]byte, 8)
	rand.Read(b)
	ans, _ := binary.Uvarint(b)
	return uint(ans)
}

// 0th position is the most significant bit
// True if bit is 1 and False if bit is 0
// N is the number of bits in uint
func getBit(n, pos, N uint) byte {
	return byte((n & (1 << (N - pos))) >> (N - pos))
}

// fixed key PRF (Matyas–Meyer–Oseas one way compression function)
// numBlocks represents the number
func prf(x []byte, aesBlocks []cipher.Block, numBlocks uint, temp, out []byte) {
	// If request blocks greater than actual needed blocks, grow output array
	if numBlocks > initPRFLen {
		out = make([]byte, numBlocks*aes.BlockSize)
	}
	for i := uint(0); i < numBlocks; i++ {
		// get AES_k[i](x)
		aesBlocks[i].Encrypt(temp, x)
		// get AES_k[i](x) ^ x
		for j := range temp {
			out[i*aes.BlockSize+uint(j)] = temp[j] ^ x[j]
		}
	}
}

// TODO: fix comment
// fixed key PRF (Matyas–Meyer–Oseas one way compression function)
// numBlocks represents the number
func prfKey(x []byte, key [][]byte, numBlocks uint, temp, out []byte) {
	// If request blocks greater than actual needed blocks, grow output array
	if numBlocks > initPRFLen {
		out = make([]byte, numBlocks*aes.BlockSize)
	}
	for i := uint(0); i < numBlocks; i++ {
		// get AES_k[i](x)
		currentKey := key[i]
		nk := len(currentKey) / 4
		k := make([]uint32, nk)
		for j := 0; j < nk; i++ {
			k[j] = binary.BigEndian.Uint32(currentKey[4*j:])
		}
		aes128MMO(&k[0], &temp[0], &x[0])
		//aesBlocks[i].Encrypt(temp, x)
		// get AES_k[i](x) ^ x
		for j := range temp {
			out[i*aes.BlockSize+uint(j)] = temp[j] ^ x[j]
		}

	}
}
