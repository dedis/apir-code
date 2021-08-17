package fss

// This file contains the server side code for the FSS library.
// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/server.go

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
)

// Upon receiving query from client, initialize server with
// this function. The server, unlike the client
// receives prfKeys, so it doesn't need to pick random ones
func ServerInitialize(prfKeys [][]byte, numBits uint) *Fss {
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
	f.N = field.Bytes * 8
	f.Temp = make([]byte, aes.BlockSize)
	f.Out = make([]byte, aes.BlockSize*initPRFLen)

	return f
}

func (f Fss) EvaluatePF(serverNum byte, k FssKeyEq2P, x uint, out []uint32) {
	sCurr := make([]byte, aes.BlockSize)
	copy(sCurr, k.SInit)
	tCurr := k.TInit
	for i := uint(0); i < f.NumBits; i++ {
		var xBit byte = 0
		if i != f.N {
			xBit = byte(getBit(x, (f.N - f.NumBits + i + 1), f.N))
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

	outLen := uint(len(out))

	// convert block
	tmp := make([]uint32, outLen)
	convertBlock(f, sCurr, tmp)
	for i := range out {
		if serverNum == 0 {
			// tCurr is either 0 or 1, no need to mod
			out[i] = (tmp[i] + uint32(tCurr)*k.FinalCW[i]) % constants.ModP
		} else {
			out[i] = constants.ModP - ((tmp[i] + uint32(tCurr)*k.FinalCW[i]) % constants.ModP)
		}
	}
}

// This is the 2-party FSS evaluation function for interval functions, i.e. <,> functions.
// The usage is similar to 2-party FSS for equality functions
// TODO: need to change the signature of the function and input a out vector
// to have the correct length
func (f Fss) EvaluateLt(k ServerKeyLt, x uint32) []uint32 {
	xBit := getBit(uint(x), (f.N - f.NumBits + 1), f.N)
	s := make([]byte, aes.BlockSize)
	copy(s, k.s[xBit])
	t := k.t[xBit]
	v := k.v[xBit]
	for i := uint(1); i < f.NumBits; i++ {
		// Get current bit
		if i != f.N {
			xBit = getBit(uint(x), uint(f.N-f.NumBits+i+1), f.N)
		} else {
			xBit = 0
		}

		// TODO: check this until end of function
		// we need two blocks for s1, one entire block for both bits in t1 and
		// for each of the two vectors, we need blockLength/4 (i.e., len(v0[0])/4)
		// so len(v0[0])/2
		numBlocks := uint(2 + 1 + len(v)/field.Bytes)
		prf(s, f.FixedBlocks, numBlocks, f.Temp, f.Out)

		// Pick the right values to use based on bit of x
		xStart := int(aes.BlockSize * xBit)
		copy(s, f.Out[xStart:xStart+aes.BlockSize])

		for j := 0; j < aes.BlockSize; j++ {
			s[j] = s[j] ^ k.cw[t][i-1].cs[xBit][j]
		}
		vStart := aes.BlockSize*2 + 8 + 8*xBit
		// TODO: conv should be a vector of field elements
		for i := range v {
			conv := binary.BigEndian.Uint32(f.Out[int(vStart)+4*i:int(vStart)+4*(i+1)]) % field.ModP
			v[i] = (v[i] + conv + k.cw[t][i-1].cv[xBit][i]) % field.ModP
		}
		t = (uint8(f.Out[2*aes.BlockSize+xBit]) % 2) ^ k.cw[t][i-1].ct[xBit]
	}

	return v
}
