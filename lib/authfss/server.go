package authfss

// This file contains the server side code for authenticated the FSS library.
// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/server.go

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/si-co/vpir-code/lib/field"
)

// Upon receiving query from client, initialize server with
// this function. The server, unlike the client
// receives prfKeys, so it doesn't need to pick random ones
func ServerInitialize(blockLength int) *Fss {
	f := new(Fss)
	f.FixedBlocks = make([]cipher.Block, len(PrfKeys))
	for i := range PrfKeys {
		block, err := aes.NewCipher(PrfKeys[i])
		if err != nil {
			panic(err.Error())
		}
		f.FixedBlocks[i] = block
	}
	f.N = 256 // maximum number of bits supported by FSS
	f.Temp = make([]byte, aes.BlockSize)
	f.Out = make([]byte, aes.BlockSize*len(PrfKeys))
	f.OutConvertBlock = make([]byte, blockLength*field.Bytes)

	return f
}

func (f Fss) EvaluatePF(serverNum byte, k FssKeyEq2P, x []bool, out []uint32) {
	// reinitialize f.NumBits because we have different input lengths
	f.NumBits = uint(len(x))

	sCurr := make([]byte, aes.BlockSize)
	copy(sCurr, k.SInit)
	tCurr := k.TInit
	tmp := make([]uint32, len(out))
	for i := uint(0); i < f.NumBits; i++ {
		var xBit byte = 0
		if i != f.N {
			// original: xBit = byte(getBit(x, (f.N - f.NumBits + i + 1), f.N))
			if x[i] {
				xBit = 1
			}
		}

		prf(sCurr, f.FixedBlocks, 3, f.Temp, f.Out)

		// Keep counter to ensure we are accessing CW correctly
		count := 0
		for j := 0; j < aes.BlockSize*2+2; j++ {
			// Make sure we are doing G(s) ^ (t*sCW||tLCW||sCW||tRCW)
			if j == aes.BlockSize+1 {
				count = 0
			} else if j == aes.BlockSize*2+1 {
				count = aes.BlockSize + 1
			}
			f.Out[j] = f.Out[j] ^ (tCurr * k.CW[i][count])
			count++
		}

		// Pick right seed expansion based on
		if xBit == 0 {
			copy(sCurr, f.Out[:aes.BlockSize])
			tCurr = f.Out[aes.BlockSize] % 2
		} else {
			copy(sCurr, f.Out[(aes.BlockSize+1):(aes.BlockSize*2+1)])
			tCurr = f.Out[aes.BlockSize*2+1] % 2
		}
	}

	// convert block
	convertBlock(f, sCurr, tmp)
	for i := range out {
		if serverNum == 0 {
			// tCurr is either 0 or 1, no need to mod
			out[i] = (tmp[i] + uint32(tCurr)*k.FinalCW[i]) % field.ModP
		} else {
			out[i] = field.ModP - ((tmp[i] + uint32(tCurr)*k.FinalCW[i]) % field.ModP)
		}
	}
}
