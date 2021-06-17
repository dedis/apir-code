package utils

import (
	"math"
	"math/bits"
)

func BytesTobits(bs []byte) []byte {
	r := make([]byte, len(bs)*8)
	for i, b := range bs {
		for j := 0; j < 8; j++ {
			r[i*8+j] = byte(b >> uint(7-j) & 0x01)
		}
	}
	return r
}

func BitsPadded(bs []byte) []byte {
	st := make([]byte, len(bs)*8)
	for i, d := range bs {
		for j := 0; j < 8; j++ {
			if bits.LeadingZeros8(d) == 0 {
				// No leading 0 means that it is a 1
				st[i*8+j] = 1
			} else {
				st[i*8+j] = 0
			}
			d = d << 1
		}
	}

	out := make([]byte, 0)
	for j := 0; j < len(st); j += 7 {
		end := j + 7
		if end > len(st) {
			end = len(st)
		}
		this := append([]byte{0}, st[j:end]...)
		out = append(out, this...)
	}
	return out
}

func PackBits(bits []byte) []byte {
	out := make([]byte, int(math.Ceil(float64(len(bits))/float64(8))))
	for i := 0; i < len(bits); i++ {
		out[i/8] |= bits[i] << (7 - (i % 8))
	}

	return out
}

func BytesUnpadded(bits []byte) []byte {
	// remove zeros at beginning
	for i := len(bits) - 1; i >= 0; i-- {
		if bits[i] == 0 {
			bits = bits[:i]
		} else {
			break
		}
	}

	out := make([]byte, 0)
	for i := 0; i < len(bits); i++ {
		if i%8 == 0 {
			continue
		}
		out = append(out, bits[i])
	}

	out = PackBits(out)

	return out
}
