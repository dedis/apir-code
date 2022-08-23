package ecc

// ECC defines the parameters used for the error correcting code (ECC)
type ECC struct {
	t int
}

func (e *ECC) Encode(in []byte) []byte {
	cwLength := 2 * e.t // length of codeword
	out := make([]byte, 2*e.t*len(in))
	for i := range in {
		for j := i * cwLength; j < (i+1)*cwLength; j++ {
			out[j] = in[i]
		}
	}

	return out
}

func (e *ECC) Decode(in []byte) ([]byte, error) {
	panic("not yet implemented")
}
