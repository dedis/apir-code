package authfss

// This file contains all the client code for the authenticated FSS scheme.
// Source: https://github.com/frankw2/libfss/blob/master/go/libfss/client.go

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/si-co/vpir-code/lib/field"
)

// Initialize client with this function
// numBits represents the input domain for the function, i.e. the number
// of bits to check
// TODO: initialize PRF keys as in Dima's init() function, so that we don't
// have to send the prfKeys to the server and we avoid initialization of PRF
// keys every time we invoke the client. We can use the expand key function to
// and replicate the same behaviour as the init function in the other library
func ClientInitialize(blockLength int) *Fss {
	f := new(Fss)
	f.BlockLength = blockLength
	initPRFLen := 4
	f.PrfKeys = make([][]byte, initPRFLen)
	// Create fixed AES blocks
	f.FixedBlocks = make([]cipher.Block, initPRFLen)
	for i := uint(0); i < uint(initPRFLen); i++ {
		f.PrfKeys[i] = make([]byte, aes.BlockSize)
		rand.Read(f.PrfKeys[i])
		block, err := aes.NewCipher(f.PrfKeys[i])
		if err != nil {
			panic(err.Error())
		}
		f.FixedBlocks[i] = block
	}
	f.N = 256 // maximum number of bits supported by FSS
	f.Temp = make([]byte, aes.BlockSize)
	f.Out = make([]byte, aes.BlockSize*initPRFLen)
	f.OutConvertBlock = make([]byte, blockLength*field.Bytes)

	return f
}

// Generate Keys for 2-party point functions It creates keys for a function
// that evaluates to vector b when input x = a.
func (f Fss) GenerateTreePF(a []bool, b []uint32) []FssKeyEq2P {
	// reinitialize f.NumBits because we have different input lengths
	f.NumBits = uint(len(a))

	fssKeys := make([]FssKeyEq2P, 2)
	// Set up initial values
	tempRand1 := make([]byte, aes.BlockSize+1)
	rand.Read(tempRand1)
	fssKeys[0].SInit = tempRand1[:aes.BlockSize]
	fssKeys[0].TInit = tempRand1[aes.BlockSize] % 2
	fssKeys[1].SInit = make([]byte, aes.BlockSize)
	rand.Read(fssKeys[1].SInit)
	fssKeys[1].TInit = fssKeys[0].TInit ^ 1

	// Set current seed being used
	sCurr0 := make([]byte, aes.BlockSize)
	sCurr1 := make([]byte, aes.BlockSize)
	copy(sCurr0, fssKeys[0].SInit)
	copy(sCurr1, fssKeys[1].SInit)
	tCurr0 := fssKeys[0].TInit
	tCurr1 := fssKeys[1].TInit

	// Initialize correction words in FSS keys
	fssKeys[0].CW = make([][]byte, f.NumBits)
	fssKeys[1].CW = make([][]byte, f.NumBits)
	for i := uint(0); i < f.NumBits; i++ {
		// make AES block size + 2 bytes
		fssKeys[0].CW[i] = make([]byte, aes.BlockSize+2)
		fssKeys[1].CW[i] = make([]byte, aes.BlockSize+2)
	}

	leftStart := 0
	rightStart := aes.BlockSize + 1
	for i := uint(0); i < f.NumBits; i++ {
		// "expand" seed into two seeds + 2 bits
		prf(sCurr0, f.FixedBlocks, 3, f.Temp, f.Out)
		prfOut0 := make([]byte, aes.BlockSize*3)
		copy(prfOut0, f.Out[:aes.BlockSize*3])
		prf(sCurr1, f.FixedBlocks, 3, f.Temp, f.Out)
		prfOut1 := make([]byte, aes.BlockSize*3)
		copy(prfOut1, f.Out[:aes.BlockSize*3])

		// Parse out "t" bits
		t0Left := prfOut0[aes.BlockSize] % 2
		t0Right := prfOut0[(aes.BlockSize*2)+1] % 2
		t1Left := prfOut1[aes.BlockSize] % 2
		t1Right := prfOut1[(aes.BlockSize*2)+1] % 2
		// Find bit in a
		// original: aBit := getBit(a, (f.N - f.NumBits + i + 1), f.N)
		aBit := byte(0)
		if a[i] {
			aBit = byte(1)
		}

		// Figure out which half of expanded seeds to keep and lose
		keep := rightStart
		lose := leftStart
		if aBit == 0 {
			keep = leftStart
			lose = rightStart
		}

		// Set correction words for both keys. Note: they are the same
		for j := 0; j < aes.BlockSize; j++ {
			fssKeys[0].CW[i][j] = prfOut0[lose+j] ^ prfOut1[lose+j]
			fssKeys[1].CW[i][j] = fssKeys[0].CW[i][j]
		}
		fssKeys[0].CW[i][aes.BlockSize] = t0Left ^ t1Left ^ aBit ^ 1
		fssKeys[1].CW[i][aes.BlockSize] = fssKeys[0].CW[i][aes.BlockSize]
		fssKeys[0].CW[i][aes.BlockSize+1] = t0Right ^ t1Right ^ aBit
		fssKeys[1].CW[i][aes.BlockSize+1] = fssKeys[0].CW[i][aes.BlockSize+1]

		for j := 0; j < aes.BlockSize; j++ {
			sCurr0[j] = prfOut0[keep+j] ^ (tCurr0 * fssKeys[0].CW[i][j])
			sCurr1[j] = prfOut1[keep+j] ^ (tCurr1 * fssKeys[0].CW[i][j])
		}

		tCWKeep := fssKeys[0].CW[i][aes.BlockSize]
		if keep == rightStart {
			tCWKeep = fssKeys[0].CW[i][aes.BlockSize+1]
		}
		tCurr0 = (prfOut0[keep+aes.BlockSize] % 2) ^ tCWKeep*tCurr0
		tCurr1 = (prfOut1[keep+aes.BlockSize] % 2) ^ tCWKeep*tCurr1
	}

	bLen := uint(len(b))

	// convert blocks
	tmp0 := make([]uint32, bLen)
	tmp1 := make([]uint32, bLen)
	convertBlock(f, sCurr0, tmp0)
	convertBlock(f, sCurr1, tmp1)

	fssKeys[0].FinalCW = make([]uint32, bLen)
	fssKeys[1].FinalCW = make([]uint32, bLen)

	for i := range fssKeys[0].FinalCW {
		// Need to make sure that no intermediate
		// results under or overflow the 32-bit modulus

		//fssKeys[0].FinalCW[i] = (b[i] - tmp0[i] + tmp1[i]) % field.ModP
		val := (b[i] + (field.ModP - tmp0[i])) % field.ModP
		val = (val + tmp1[i]) % field.ModP
		fssKeys[0].FinalCW[i] = val
		fssKeys[1].FinalCW[i] = fssKeys[0].FinalCW[i]
		if tCurr1 == 1 {
			fssKeys[0].FinalCW[i] = field.ModP - fssKeys[0].FinalCW[i] // negation
			fssKeys[1].FinalCW[i] = fssKeys[0].FinalCW[i]
		}
	}

	return fssKeys
}
