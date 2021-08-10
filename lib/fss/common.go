package libfss

// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/common.go

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
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

// Structs for keys
type FssKeyEq2P struct {
	SInit   []byte
	TInit   byte
	CW      [][]byte // there are n
	FinalCW int
}

type FssKeyEq2PVector struct {
	SInit   []byte
	TInit   byte
	CW      [][]byte // there are n
	FinalCW []int
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
		// generate new key if needed
		if i < uint(len(aesBlocks)) {
			// get AES_k[i](x)
			aesBlocks[i].Encrypt(temp, x)
		} else {
			// TODO: generate and store a new AES key?
			prfKey := make([]byte, aes.BlockSize)
			rand.Read(prfKey)
			block, err := aes.NewCipher(prfKey)
			if err != nil {
				panic(err.Error())
			}
			block.Encrypt(temp, x)
		}

		// get AES_k[i](x) ^ x
		for j := range temp {
			out[i*aes.BlockSize+uint(j)] = temp[j] ^ x[j]
		}
	}
}

func convertBlock(f Fss, x []byte, out []int64) {
	bLen := uint(len(out))

	randBytes := make([]byte, bLen*8)

	// we can generate two int per AES block since they are supposed to be
	// 64 bits
	prf(x, f.FixedBlocks, bLen/2, f.Temp, randBytes)

	for i := range out {
		out[i], _ = binary.Varint(randBytes[i*8 : (i+1)*8])
	}
}
