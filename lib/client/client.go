package client

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Client represents the client instance in both the IT and DPF-based schemes
type Client interface {
	QueryBytes(int, int) ([][]byte, error)
	ReconstructBytes([][]byte) ([]field.Element, error)
}

type state struct {
	ix    int
	iy    int
	alpha field.Element
	a     []field.Element
}

// general functions for both IT and DPF-based clients
func decodeAnswer(a [][]byte) ([][][]field.Element, error) {
	// servers answers
	answer := make([][][]field.Element, len(a))
	for i, ans := range a {
		buf := bytes.NewBuffer(ans)
		dec := gob.NewDecoder(buf)
		var serverAnswer [][]field.Element
		if err := dec.Decode(&serverAnswer); err != nil {
			return nil, err
		}
		answer[i] = serverAnswer
	}

	return answer, nil
}

func encodeReconstruct(r []field.Element) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generateClientState(index int, rnd io.Reader, dbInfo *database.Info) (*state, error) {
	// initialize state
	st := &state{}

	// sample random alpha using blake2b
	if _, err := st.alpha.SetRandom(rnd); err != nil {
		return nil, err
	}

	// Compute the position in the db (vector or matrix)
	// if db is a vector, ix always equals 0
	st.ix = index / dbInfo.NumColumns
	st.iy = index % dbInfo.NumColumns

	if dbInfo.BlockSize != cst.SingleBitBlockLength {
		// compute vector a = (1, alpha, alpha^2, ..., alpha^b) for the
		// multi-bit scheme
		// +1 to BlockSize for recovering true value
		st.a = make([]field.Element, dbInfo.BlockSize+1)
		st.a[0] = field.One()
		st.a[1] = st.alpha
		for i := 2; i < len(st.a); i++ {
			st.a[i].Mul(&st.a[i-1], &st.alpha)
		}
	} else {
		// the single-bit scheme needs a single alpha
		st.a = make([]field.Element, 1)
		st.a[0] = st.alpha
	}

	return st, nil
}

func reconstruct(answers [][][]field.Element, dbInfo *database.Info, st *state) ([]field.Element, error) {
	sum := make([][]field.Element, dbInfo.NumRows)

	if dbInfo.BlockSize == cst.SingleBitBlockLength {
		// sum answers as vectors in F^b
		for i := 0; i < dbInfo.NumRows; i++ {
			sum[i] = make([]field.Element, 1)
			for k := range answers {
				sum[i][0].Add(&sum[i][0], &answers[k][i][0])
			}
		}
		for i := 0; i < dbInfo.NumRows; i++ {
			if i == st.ix {
				switch {
				case sum[i][0].Equal(&st.alpha):
					return []field.Element{cst.One}, nil
				case sum[i][0].Equal(&cst.Zero):
					return []field.Element{cst.Zero}, nil
				default:
					return nil, errors.New("REJECT!")
				}
			} else {
				if !sum[i][0].Equal(&st.alpha) && !sum[i][0].Equal(&cst.Zero) {
					return nil, errors.New("REJECT!")
				}
			}
		}
	}

	// sum answers as vectors in F^(b+1)
	for i := 0; i < dbInfo.NumRows; i++ {
		sum[i] = make([]field.Element, dbInfo.BlockSize+1)
		for b := 0; b < dbInfo.BlockSize+1; b++ {
			for k := range answers {
				sum[i][b].Add(&sum[i][b], &answers[k][i][b])
			}
		}
	}
	var tag, prod field.Element
	messages := make([]field.Element, dbInfo.BlockSize)
	for i := 0; i < dbInfo.NumRows; i++ {
		copy(messages, sum[i][:len(sum[i])-1])
		tag = sum[i][len(sum[i])-1]
		// compute reconstructed tag
		reconstructedTag := field.Zero()
		for b := 0; b < len(messages); b++ {
			prod.Mul(&st.a[b+1], &messages[b])
			reconstructedTag.Add(&reconstructedTag, &prod)
		}
		if !tag.Equal(&reconstructedTag) {
			fmt.Println("tag:", tag)
			fmt.Println("rec:", reconstructedTag)
			//return nil, errors.New("REJECT")
		}
	}

	return sum[st.ix][:len(sum[st.ix])-1], nil
}

// return true if the query inputs are invalid for IT schemes
func invalidQueryInputsIT(index, numServers int) bool {
	return index < 0 && numServers < 2
}

// return true if the query inputs are invalid for DPF-based schemes
func invalidQueryInputsDPF(index, numServers int) bool {
	return index < 0 && numServers != 2
}
