package fss

// This file contains all the client code for the FSS scheme.
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
func ClientInitialize(numBits uint, blockLength int) *Fss {
	f := new(Fss)
	f.NumBits = numBits
	initPRFLen := blockLength / field.Bytes
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
	f.N = field.Bytes * 8
	f.Temp = make([]byte, aes.BlockSize)
	f.Out = make([]byte, aes.BlockSize*initPRFLen)
	return f
}

// Generate Keys for 2-party point functions It creates keys for a function
// that evaluates to vector b when input x = a.
func (f Fss) GenerateTreePF(a uint32, b []uint32) []FssKeyEq2P {
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
		aBit := getBit(uint(a), (f.N - f.NumBits + i + 1), f.N)

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
	convertBlock(f, sCurr0, &tmp0)
	convertBlock(f, sCurr1, &tmp1)

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

// This function contains the 2-party FSS key generation for interval
// functions, i.e. <, > functions.  The usage is similar to 2-party FSS for
// equality functions.
// From: Boyle et al., Function Secret Sharing, EUROCRYPT'15, pag. 19
func (f Fss) GenerateTreeLt(a uint32, b []uint32) []ServerKeyLt {
	lenb := len(b)

	k := make([]ServerKeyLt, 2)

	k[0].cw = make([][]CWLt, 2)
	k[0].cw[0] = make([]CWLt, f.NumBits)
	k[0].cw[1] = make([]CWLt, f.NumBits)
	k[1].cw = make([][]CWLt, 2)
	k[1].cw[0] = make([]CWLt, f.NumBits)
	k[1].cw[1] = make([]CWLt, f.NumBits)

	k[0].s = make([][]byte, 2)
	k[0].s[0] = make([]byte, aes.BlockSize)
	k[0].s[1] = make([]byte, aes.BlockSize)
	k[1].s = make([][]byte, 2)
	k[1].s[0] = make([]byte, aes.BlockSize)
	k[1].s[1] = make([]byte, aes.BlockSize)

	k[0].t = make([]uint8, 2)
	k[1].t = make([]uint8, 2)
	k[0].v = make([][]uint32, 2)
	k[1].v = make([][]uint32, 2)
	// Figure out first bit
	aBit := getBit(uint(a), (f.N - f.NumBits + 1), f.N)
	naBit := aBit ^ 1

	// Initialize seeds (store as an array for each server)
	// The first AES_SIZE bits are for the 0 bit
	// The second AES_SIZE bits are for the 1 bit
	s0 := make([]byte, aes.BlockSize*2)
	s1 := make([]byte, aes.BlockSize*2)
	aStart := int(aes.BlockSize * aBit)
	naStart := int(aes.BlockSize * naBit)

	rand.Read(s0[aStart : aStart+aes.BlockSize])
	rand.Read(s1[aStart : aStart+aes.BlockSize])
	rand.Read(s0[naStart : naStart+aes.BlockSize])
	// Ensure the "not a" bits are the same
	copy(s1[naStart:naStart+aes.BlockSize], s0[naStart:naStart+aes.BlockSize])

	// Set initial "t" bits
	t0 := make([]uint8, 2)
	t1 := make([]uint8, 2)
	temp := make([]byte, 2)
	rand.Read(temp)

	// Make sure t0a and t1a are different
	t0[aBit] = uint8(temp[0]) % 2
	t1[aBit] = t0[aBit] ^ 1

	// Make sure t0na = t1na
	t0[naBit] = uint8(temp[1]) % 2
	t1[naBit] = t0[naBit]

	// Generate random Vs
	// NOTE: element of the output group, which for us is F^(1+b)
	v0 := make([][]uint32, 2)
	v1 := make([][]uint32, 2)

	// make sure v0a + -v1a = 0
	v0[aBit] = field.RandVector(lenb)
	v1[aBit] = field.NegateVector(v0[aBit])

	// make sure v0na + -v1na = a1 * b
	// NOTE: b is the target input, denoted as g in the paper, in F^(1+b)
	v0[naBit] = field.RandVector(lenb)
	if aBit == 0 {
		v1[naBit] = field.NegateVector((v0[naBit]))
	} else {
		for i := range v1[naBit] {
			// v1[naBit][i] = v0[naBit][i] - b[i]
			v1[naBit][i] = (v0[naBit][i] + (field.ModP - b[i])) % field.ModP
		}
	}

	// Store generated values into the key
	copy(k[0].s[0], s0[0:aes.BlockSize])
	copy(k[0].s[1], s0[aes.BlockSize:aes.BlockSize*2])
	copy(k[1].s[0], s1[0:aes.BlockSize])
	copy(k[1].s[1], s1[aes.BlockSize:aes.BlockSize*2])
	k[0].t[0] = t0[0]
	k[0].t[1] = t0[1]
	k[1].t[0] = t1[0]
	k[1].t[1] = t1[1]
	k[0].v[0] = v0[0]
	k[0].v[1] = v0[1]
	k[1].v[0] = v1[0]
	k[1].v[1] = v1[1]

	// Assign keys and start cipher
	key0 := make([]byte, aes.BlockSize)
	key1 := make([]byte, aes.BlockSize)
	copy(key0, s0[aStart:aStart+aes.BlockSize])
	copy(key1, s1[aStart:aStart+aes.BlockSize])
	tbit0, tbit1 := t0[aBit], t1[aBit]

	cs0 := make([]byte, aes.BlockSize*2)
	cs1 := make([]byte, aes.BlockSize*2)
	ct0 := make([]uint8, 2)
	ct1 := make([]uint8, 2)

	// TODO: elements of the group
	var cv [][][]uint32
	cv = make([][][]uint32, 2)
	cv[0] = make([][]uint32, 2)
	cv[1] = make([][]uint32, 2)

	for i := uint(0); i < f.NumBits-1; i++ {
		// Figure out next bit
		aBit = getBit(uint(a), (f.N - f.NumBits + i + 2), f.N)
		naBit = aBit ^ 1

		// TODO: check this until line 288
		prf(key0, f.FixedBlocks, 3, f.Temp, f.Out)
		copy(s0, f.Out[:aes.BlockSize*2])    // 2 blocks here
		t0[0] = f.Out[aes.BlockSize*2] % 2   // one byte from third block
		t0[1] = f.Out[aes.BlockSize*2+1] % 2 // one additional byte from third block
		// TODO: here we are wasting 14 bytes from the third block

		convertBlock(f, s0, &(v0[0]))
		convertBlock(f, s0, &(v0[1]))

		prf(key1, f.FixedBlocks, 3, f.Temp, f.Out)
		copy(s1, f.Out[:aes.BlockSize*2])
		t1[0] = f.Out[aes.BlockSize*2] % 2
		t1[1] = f.Out[aes.BlockSize*2+1] % 2

		convertBlock(f, s1, &(v1[0]))
		convertBlock(f, s1, &(v1[1]))

		// Redefine aStart and naStart based on new a's
		aStart = int(aes.BlockSize * aBit)
		naStart = int(aes.BlockSize * naBit)

		// Create cs and ct for next bit
		rand.Read(cs0[aStart : aStart+aes.BlockSize])
		rand.Read(cs1[aStart : aStart+aes.BlockSize])

		// Pick random cs0na and pick cs1na s.t.
		// cs0na xor cs1na xor s0na xor s1na = 0
		rand.Read(cs0[naStart : naStart+aes.BlockSize])

		for j := 0; j < aes.BlockSize; j++ {
			cs1[naStart+j] = s0[naStart+j] ^ s1[naStart+j] ^ cs0[naStart+j]
		}

		rand.Read(temp)
		// Set ct0a and ct1a s.t.
		// ct0a xor ct1a xor t0a xor t1a = 1
		ct0[aBit] = uint8(temp[0]) % 2
		ct1[aBit] = ct0[aBit] ^ t0[aBit] ^ t1[aBit] ^ 1

		// Set ct0na and ct1na s.t.
		// ct0na xor ct1na xor t0na xor t1na = 0
		ct0[naBit] = uint8(temp[1]) % 2
		ct1[naBit] = ct0[naBit] ^ t0[naBit] ^ t1[naBit]

		// NOTE: cv are group element
		cv[tbit0][aBit] = field.RandVector(lenb)
		cv[tbit1][aBit] = make([]uint32, lenb)
		for i := range cv[tbit0][aBit] {
			//cv[tbit1][aBit][i] = (v0[aBit][i] + cv[tbit0][aBit][i] - v1[aBit][i]) % field.ModP
			val := (v0[aBit][i] + cv[tbit0][aBit][i]) % field.ModP
			val = (val + (field.ModP - v1[aBit][i])) % field.ModP
			cv[tbit1][aBit][i] = val
		}

		cv[tbit0][naBit] = field.RandVector(lenb)
		cv[tbit1][naBit] = make([]uint32, lenb)
		for i := range cv[tbit0][naBit] {
			//cv[tbit1][naBit][i] = (cv[tbit0][naBit][i] +
			//v0[naBit][i] -
			//v1[naBit][i] -
			//b[i]*uint32(aBit)) % field.ModP
			val := (cv[tbit0][naBit][i] + v0[naBit][i]) % field.ModP
			val = (val + (field.ModP - v1[naBit][i])) % field.ModP
			val = (val + (field.ModP - b[i]*uint32(aBit))) % field.ModP
			cv[tbit1][naBit][i] = val
		}

		k[0].cw[0][i].cs = make([][]byte, 2)
		k[0].cw[0][i].cs[0] = make([]byte, aes.BlockSize)
		k[0].cw[0][i].cs[1] = make([]byte, aes.BlockSize)
		k[0].cw[1][i].cs = make([][]byte, 2)
		k[0].cw[1][i].cs[0] = make([]byte, aes.BlockSize)
		k[0].cw[1][i].cs[1] = make([]byte, aes.BlockSize)

		k[0].cw[0][i].ct = make([]uint8, 2)
		k[0].cw[0][i].cv = make([][]uint32, 2) // NOTE: group element
		k[0].cw[1][i].ct = make([]uint8, 2)
		k[0].cw[1][i].cv = make([][]uint32, 2) // NOTE: group element

		copy(k[0].cw[0][i].cs[0], cs0[0:aes.BlockSize])
		copy(k[0].cw[0][i].cs[1], cs0[aes.BlockSize:aes.BlockSize*2])
		k[0].cw[0][i].ct[0] = ct0[0]
		k[0].cw[0][i].ct[1] = ct0[1]
		copy(k[0].cw[1][i].cs[0], cs1[0:aes.BlockSize])
		copy(k[0].cw[1][i].cs[1], cs1[aes.BlockSize:aes.BlockSize*2])
		k[0].cw[1][i].ct[0] = ct1[0]
		k[0].cw[1][i].ct[1] = ct1[1]

		k[0].cw[0][i].cv[0] = cv[0][0]
		k[0].cw[0][i].cv[1] = cv[0][1]
		k[0].cw[1][i].cv[0] = cv[1][0]
		k[0].cw[1][i].cv[1] = cv[1][1]

		k[1].cw[0][i].cs = make([][]byte, 2)
		k[1].cw[0][i].cs[0] = make([]byte, aes.BlockSize)
		k[1].cw[0][i].cs[1] = make([]byte, aes.BlockSize)
		k[1].cw[1][i].cs = make([][]byte, 2)
		k[1].cw[1][i].cs[0] = make([]byte, aes.BlockSize)
		k[1].cw[1][i].cs[1] = make([]byte, aes.BlockSize)

		k[1].cw[0][i].ct = make([]uint8, 2)
		k[1].cw[0][i].cv = make([][]uint32, 2)
		k[1].cw[1][i].ct = make([]uint8, 2)
		k[1].cw[1][i].cv = make([][]uint32, 2)

		copy(k[1].cw[0][i].cs[0], cs0[0:aes.BlockSize])
		copy(k[1].cw[0][i].cs[1], cs0[aes.BlockSize:aes.BlockSize*2])
		k[1].cw[0][i].ct[0] = ct0[0]
		k[1].cw[0][i].ct[1] = ct0[1]
		copy(k[1].cw[1][i].cs[0], cs1[0:aes.BlockSize])
		copy(k[1].cw[1][i].cs[1], cs1[aes.BlockSize:aes.BlockSize*2])
		k[1].cw[1][i].ct[0] = ct1[0]
		k[1].cw[1][i].ct[1] = ct1[1]

		k[1].cw[0][i].cv[0] = cv[0][0]
		k[1].cw[0][i].cv[1] = cv[0][1]
		k[1].cw[1][i].cv[0] = cv[1][0]
		k[1].cw[1][i].cv[1] = cv[1][1]

		// Find correct cs and ct
		var cs, ct []byte

		// Set next seeds and ts
		if tbit0 == 1 {
			cs = cs1
			ct = ct1
		} else {
			cs = cs0
			ct = ct0
		}
		for j := 0; j < len(key0); j++ {
			key0[j] = s0[aStart+j] ^ cs[aStart+j]
		}
		tbit0 = t0[aBit] ^ ct[aBit]
		if tbit1 == 1 {
			cs = cs1
			ct = ct1
		} else {
			cs = cs0
			ct = ct0
		}

		for j := 0; j < len(key1); j++ {
			key1[j] = s1[aStart+j] ^ cs[aStart+j]
		}

		tbit1 = t1[aBit] ^ ct[aBit]
	}

	return k
}
