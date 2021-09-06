package fss

// This file contains the server side code for the FSS library.
// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/server.go

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/si-co/vpir-code/lib/field"
)

// Upon receiving query from client, initialize server with
// this function. The server, unlike the client
// receives prfKeys, so it doesn't need to pick random ones
func ServerInitialize(prfKeys [][]byte, numBits uint, blockLength int) *Fss {
	f := new(Fss)
	initPRFLen := len(prfKeys)
	f.NumBits = numBits
	f.PrfKeys = make([][]byte, initPRFLen)
	f.FixedBlocks = make([]cipher.Block, initPRFLen)
	for i := range prfKeys {
		f.PrfKeys[i] = make([]byte, aes.BlockSize)
		copy(f.PrfKeys[i], prfKeys[i])
		block, err := aes.NewCipher(f.PrfKeys[i])
		if err != nil {
			panic(err.Error())
		}
		f.FixedBlocks[i] = block
	}
	f.N = 256 // maximum number of bits supported by FSS
	f.Temp = make([]byte, aes.BlockSize)
	f.Out = make([]byte, aes.BlockSize*initPRFLen)
	f.OutConvertBlock = make([]byte, (blockLength+1)*field.Bytes)

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

// This is the 2-party FSS evaluation function for interval functions, i.e. <,> functions.
// The usage is similar to 2-party FSS for equality functions
func (f Fss) EvaluateLt(k ServerKeyLt, x uint64) []uint32 {
	xBit := getBit(x, (f.N - f.NumBits + 1), f.N)
	s := make([]byte, aes.BlockSize)
	copy(s, k.s[xBit])
	t := k.t[xBit]
	v := k.v[xBit]
	for i := uint(1); i < f.NumBits; i++ {
		// Get current bit
		if i != f.N {
			xBit = getBit(x, uint(f.N-f.NumBits+i+1), f.N)
		} else {
			xBit = 0
		}

		// TODO: check this until end of function
		prf(s, f.FixedBlocks, 4, f.Temp, f.Out) // TODO: we are waisting a block?

		// Pick the right values to use based on bit of x
		xStart := int(aes.BlockSize * xBit)
		copy(s, f.Out[xStart:xStart+aes.BlockSize])

		for j := 0; j < aes.BlockSize; j++ {
			s[j] = s[j] ^ k.cw[t][i-1].cs[xBit][j]
		}

		conv := make([]uint32, len(v))
		convertBlock(f, s, conv)
		for j := range v {
			//v[j] = (v[j] + conv[j] + k.cw[t][i-1].cv[xBit][j]) % field.ModP
			val := (v[j] + conv[j]) % field.ModP
			val = (val + k.cw[t][i-1].cv[xBit][j]) % field.ModP
			v[j] = val
		}
		t = (uint8(f.Out[2*aes.BlockSize+xBit]) % 2) ^ k.cw[t][i-1].ct[xBit]
	}

	return v
}
