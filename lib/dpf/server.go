package dpf

// This file contains the server side code for the FSS library.

import (
	"crypto/aes"
	"crypto/cipher"

	//"fmt"

	"github.com/si-co/vpir-code/lib/field"
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
		//	fmt.Println("server")
		//	fmt.Println(f.PrfKeys[i])
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
// Boyle, Elette, Niv Gilboa, and Yuval Ishai. "Function Secret Sharing:
// Improvements and Extensions." Proceedings of the 2016 ACM SIGSAC Conference
// on Computer and Communications Security. ACM, 2016.

// EvaluatePF is executed by each of the 2 server to evaluate their function
// share on a value. Then, the client adds the results from both servers.
func (f Fss) EvaluatePF(serverNum byte, k FssKeyEq2P, x uint) *field.Element {
	sCurr := make([]byte, aes.BlockSize)
	copy(sCurr, k.SInit)
	tCurr := k.TInit
	for i := uint(0); i < f.NumBits; i++ {
		var xBit byte = 0
		if i != f.N {
			xBit = byte(getBit(x, (f.N - f.NumBits + i + 1), f.N))
		}

		prf(sCurr, f.FixedBlocks, 3, f.Temp, f.Out)
		//fmt.Println(i, sCurr)
		//fmt.Println(i, "f.Out:", f.Out)
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
		//fmt.Println("xBit", xBit)
		// Pick right seed expansion based on
		if xBit == 0 {
			copy(sCurr, f.Out[:aes.BlockSize])
			tCurr = f.Out[aes.BlockSize] % 2
		} else {
			copy(sCurr, f.Out[(aes.BlockSize+1):(aes.BlockSize*2+1)])
			tCurr = f.Out[aes.BlockSize*2+1] % 2
		}
		//fmt.Println(f.Out)
	}
	out := new(field.Element).SetBytes(sCurr)
	if tCurr > 0 {
		out.Add(out, &k.FinalCW)
	}
	if serverNum == 1 {
		out.Neg(out)
	}

	return out
}

// EvaluatePFVector is executed by each of the 2 server to evaluate their function
// share on a value. Then, the client adds the results from both servers.
func (f Fss) EvaluatePFVector(serverNum byte, ks []FssKeyEq2P, x uint) []*field.Element {
	out := make([]*field.Element, len(ks))
	//wg := sync.WaitGroup{}
	for i, k := range ks {
		//go func(i int, serverNum byte, k FssKeyEq2P, x uint) {
		//	defer wg.Done()
		out[i] = f.EvaluatePF(serverNum, k, x)
		//}(i, serverNum, k, x)
	}
	//wg.Wait()

	return out
}
