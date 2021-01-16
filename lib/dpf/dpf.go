package dpf

import (
	"crypto/rand"

	"github.com/si-co/vpir-code/lib/field"
)

type DPFkey struct {
  ServerIdx byte
  Bytes []byte
  FinalCW []field.Element
}
type block [16]byte

type bytearr struct {
	data  []byte
	index uint64
}

var prfL *aesPrf
var prfR *aesPrf
var keyL = make([]uint32, 11*4)
var keyR = make([]uint32, 11*4)

var blockStack = make([][2]*block, 63)

func init() {
	var prfkeyL = []byte{36, 156, 50, 234, 92, 230, 49, 9, 174, 170, 205, 160, 98, 236, 29, 243}
	var prfkeyR = []byte{209, 12, 199, 173, 29, 74, 44, 128, 194, 224, 14, 44, 2, 201, 110, 28}
	var errL, errR error
	prfL, errL = newCipher(prfkeyL)
	if errL != nil {
		panic("dpf: can't init AES")
	}
	prfR, errR = newCipher(prfkeyR)
	if errR != nil {
		panic("dpf: can't init AES")
	}
	expandKeyAsm(&prfkeyL[0], &keyL[0])
	expandKeyAsm(&prfkeyR[0], &keyR[0])
	//if cpu.X86.HasSSE2 == false || cpu.X86.HasAVX2 == false {
	//	panic("we need sse2 and avx")
	//}
	for i := 0; i < 63; i++ {
		blockStack[i][0] = new(block)
		blockStack[i][1] = new(block)
	}

}

func getT(in *byte) byte {
	return *in & 1
}

func clr(in *byte) {
	*in &^= 0x1
}

func convertBlock(out []field.Element, in []byte) {
  var buf [16]byte
  for i := 0; i < len(out); i++ {
    //prfL.Encrypt(in, in)
    aes128MMO(&keyL[0], &buf[0], &in[0])
    out[i].SetBytes(buf[:])
    in[0] += 1
  }
}

func prg(seed, s0, s1 *byte) (byte, byte) {
	//prfL.Encrypt(s0, seed)
	aes128MMO(&keyL[0], s0, seed)
	t0 := getT(s0)
	clr(s0)
	//prfR.Encrypt(s1, seed)
	aes128MMO(&keyR[0], s1, seed)
	t1 := getT(s1)
	clr(s1)
	return t0, t1
}

func Gen(alpha uint64, beta []field.Element, logN uint64) (DPFkey, DPFkey) {
	if alpha >= (1<<logN) || logN > 63 {
		panic("dpf: invalid parameters")
	}
  if len(beta) > 256 {
    panic("dpf: maximum len(beta) is 256")
  }
	var ka, kb DPFkey
  ka.ServerIdx = 0
  kb.ServerIdx = 1
	var CW []byte
	s0 := new(block)
	s1 := new(block)
	scw := new(block)
	rand.Read(s0[:])
	rand.Read(s1[:])

	t0 := getT(&s0[0])
	t1 := t0 ^ 1

  betaLen := len(beta)

	clr(&s0[0])
	clr(&s1[0])

	ka.Bytes = append(ka.Bytes, s0[:]...)
	ka.Bytes = append(ka.Bytes, t0)
	kb.Bytes = append(kb.Bytes, s1[:]...)
	kb.Bytes = append(kb.Bytes, t1)

	stop := logN
	s0L := new(block)
	s0R := new(block)
	s1L := new(block)
	s1R := new(block)
	for i := uint64(0); i < stop; i++ {
		t0L, t0R := prg(&s0[0], &s0L[0], &s0R[0])
		t1L, t1R := prg(&s1[0], &s1L[0], &s1R[0])

		if (alpha & (1 << (logN - 1 - i))) != 0 {
			//KEEP = R, LOSE = L
			xor16(&scw[0], &s0L[0], &s1L[0])
			tLCW := t0L ^ t1L
			tRCW := t0R ^ t1R ^ 1
			CW = append(CW, scw[:]...)
			CW = append(CW, tLCW, tRCW)
			*s0 = *s0R
			if t0 != 0 {
				xor16(&s0[0], &s0[0], &scw[0])
			}
			*s1 = *s1R
			if t1 != 0 {
				xor16(&s1[0], &s1[0], &scw[0])
			}
			if t0 != 0 {
				t0 = t0R ^ tRCW
			} else {
				t0 = t0R
			}
			if t1 != 0 {
				t1 = t1R ^ tRCW
			} else {
				t1 = t1R
			}

		} else {
			//KEEP = L, LOSE = R
			xor16(&scw[0], &s0R[0], &s1R[0])
			tLCW := t0L ^ t1L ^ 1
			tRCW := t0R ^ t1R
			CW = append(CW, scw[:]...)
			CW = append(CW, tLCW, tRCW)
			*s0 = *s0L
			if t0 != 0 {
				xor16(&s0[0], &s0[0], &scw[0])
			}
			*s1 = *s1L
			if t1 != 0 {
				xor16(&s1[0], &s1[0], &scw[0])
			}
			if t0 != 0 {
				t0 = t0L ^ tLCW
			} else {
				t0 = t0L
			}
			if t1 != 0 {
				t1 = t1L ^ tLCW
			} else {
				t1 = t1L
			}
		}
	}

  tmp0 := make([]field.Element, betaLen)
  tmp1 := make([]field.Element, betaLen)

	convertBlock(tmp0, s0[:])
	convertBlock(tmp1, s1[:])

  ka.FinalCW = make([]field.Element, betaLen)
  kb.FinalCW = make([]field.Element, betaLen)

  for i := 0; i < betaLen; i++ {
    // FinalCW = (-1)^t . [\beta - Convert(s0) + Convert(s1)]
    ka.FinalCW[i].Sub(&beta[i], &tmp0[i])
    ka.FinalCW[i].Add(&ka.FinalCW[i], &tmp1[i])
    if t1 != 0 {
      ka.FinalCW[i].Neg(&ka.FinalCW[i])
    }

    kb.FinalCW[i].Set(&ka.FinalCW[i])
  }

	ka.Bytes = append(ka.Bytes, CW...)
	kb.Bytes = append(kb.Bytes, CW...)
	return ka, kb
}


func Eval(k DPFkey, x uint64, logN uint64, out []field.Element) {
  if len(out) != len(k.FinalCW) {
    panic("dpf: len(out) != len(k.FinalCW)")
  }

	s := new(block)
	sL := new(block)
	sR := new(block)
	copy(s[:], k.Bytes[:16])
	t := k.Bytes[16]

	stop := logN

	for i := uint64(0); i < stop; i++ {
		tL, tR := prg(&s[0], &sL[0], &sR[0])
		if t != 0 {
			sCW := k.Bytes[17+i*18 : 17+i*18+16]
			tLCW := k.Bytes[17+i*18+16]
			tRCW := k.Bytes[17+i*18+17]
			xor16(&sL[0], &sL[0], &sCW[0])
			xor16(&sR[0], &sR[0], &sCW[0])
			tL ^= tLCW
			tR ^= tRCW
		}
		if (x & (uint64(1) << (logN - 1 - i))) != 0 {
			*s = *sR
			t = tR
		} else {
			*s = *sL
			t = tL
		}
	}
	//fmt.Println("Debug", s, t)

	convertBlock(out, s[:])

  for i := 0; i < len(out); i++ {
    if t != 0 {
      out[i].Add(&out[i], &k.FinalCW[i])
    }
    if k.ServerIdx != 0 {
      out[i].Neg(&out[i])
    }
  }
}

func evalFullRecursive(k DPFkey, s *block, t byte, lvl uint64, stop uint64, index *uint64, out [][]field.Element) {
  if *index >= uint64(len(out)) {
    return
  }

	if lvl == stop {
		ss := blockStack[lvl][0]
		*ss = *s
		//aes128MMO(&keyL[0], &ss[0], &ss[0])
    if len(out[*index]) != len(k.FinalCW) {
      panic("dpf: len(out[*index]) != len(k.FinalCW)")
    }

    convertBlock(out[*index], ss[:])

    for j := 0; j < len(k.FinalCW); j++ {
      if t != 0 {
        out[*index][j].Add(&out[*index][j], &k.FinalCW[j])
      }
      if k.ServerIdx != 0 {
        out[*index][j].Neg(&out[*index][j])
      }
    }

    *index += 1
		return
	}
	sL := blockStack[lvl][0]
	sR := blockStack[lvl][1]
	tL, tR := prg(&s[0], &sL[0], &sR[0])
	if t != 0 {
		sCW := k.Bytes[17+lvl*18 : 17+lvl*18+16]
		tLCW := k.Bytes[17+lvl*18+16]
		tRCW := k.Bytes[17+lvl*18+17]
		xor16(&sL[0], &sL[0], &sCW[0])
		xor16(&sR[0], &sR[0], &sCW[0])
		tL ^= tLCW
		tR ^= tRCW
	}
	evalFullRecursive(k, sL, tL, lvl+1, stop, index, out)
	evalFullRecursive(k, sR, tR, lvl+1, stop, index, out)
}

func EvalFull(key DPFkey, logN uint64, out [][]field.Element) {
	s := new(block)
	copy(s[:], key.Bytes[:16])
	t := key.Bytes[16]
	stop := logN

  index := uint64(0)
	evalFullRecursive(key, s, t, 0, stop, &index, out)
}
