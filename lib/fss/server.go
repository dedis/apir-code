package libfss

// This file contains the server side code for the FSS library.
// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/server.go

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
)

// Upon receiving query from client, initialize server with
// this function. The server, unlike the client
// receives prfKeys, so it doesn't need to pick random ones
func ServerInitialize(prfKeys [][]byte, numBits uint) *Fss {
	f := new(Fss)
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
	// Check if int is 32 or 64 bit
	var x uint64 = 1 << 32
	if uint(x) == 0 {
		f.N = 32
	} else {
		f.N = 64
	}
	f.Temp = make([]byte, aes.BlockSize)
	f.Out = make([]byte, aes.BlockSize*initPRFLen)

	return f
}

// This is the 2-party FSS evaluation function for point functions.
// This is based on the following paper:
// Boyle, Elette, Niv Gilboa, and Yuval Ishai. "Function Secret Sharing: Improvements and Extensions." Proceedings of the 2016 ACM SIGSAC Conference on Computer and Communications Security. ACM, 2016.

// Each of the 2 server calls this function to evaluate their function
// share on a value. Then, the client adds the results from both servers.

func (f Fss) EvaluatePF(serverNum byte, k FssKeyEq2P, x uint) int {
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
	sFinal, _ := binary.Varint(sCurr[:8])
	if serverNum == 0 {
		return int(sFinal) + int(tCurr)*k.FinalCW
	} else {
		return -1 * (int(sFinal) + int(tCurr)*k.FinalCW)
	}
}

func (f Fss) EvaluatePFVector(serverNum byte, k FssKeyEq2PVector, x uint, out []int) {
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

	tmp := make([]int64, outLen)
	outPrg := make([]byte, outLen*8)

	prf(sCurr, f.FixedBlocks, outLen/2, f.Temp, outPrg)
	for i := range tmp {
		tmp[i], _ = binary.Varint(outPrg[i*8 : (i+1)*8])
	}

	for i := range out {
		if serverNum == 0 {
			out[i] = int(tmp[i]) + int(tCurr)*k.FinalCW[i]
		} else {
			out[i] = -1 * (int(tmp[i]) + int(tCurr)*k.FinalCW[i])
		}
	}
}

// This is the 2-party FSS evaluation function for interval functions, i.e. <,> functions.
// The usage is similar to 2-party FSS for equality functions
func (f Fss) EvaluateLt(k ServerKeyLt, x uint) uint {
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
		prf(s, f.FixedBlocks, 4, f.Temp, f.Out)

		// Pick the right values to use based on bit of x
		xStart := int(aes.BlockSize * xBit)
		copy(s, f.Out[xStart:xStart+aes.BlockSize])

		for j := 0; j < aes.BlockSize; j++ {
			s[j] = s[j] ^ k.cw[t][i-1].cs[xBit][j]
		}
		vStart := aes.BlockSize*2 + 8 + 8*xBit
		conv, _ := binary.Uvarint(f.Out[vStart : vStart+8])
		v = v + uint(conv) + k.cw[t][i-1].cv[xBit]
		t = (uint8(f.Out[2*aes.BlockSize+xBit]) % 2) ^ k.cw[t][i-1].ct[xBit]
	}
	return v
}
