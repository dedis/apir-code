package client

import (
	"bytes"
	"encoding/gob"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"io"

	"github.com/si-co/vpir-code/lib/field"
)

// Client represents the client instance in both the IT and DPF-based schemes
type Client interface {
	QueryBytes(int, int) ([][]byte, error)
	ReconstructBytes([]byte) ([]field.Element, error)
}

type state struct {
	ix    int
	iy    int
	alpha field.Element
	a     []field.Element
}

// general functions for both IT and DPF-based clients
func decodeAnswer(a []byte) ([][][]field.Element, error) {
	// decode answer
	buf := bytes.NewBuffer(a)
	dec := gob.NewDecoder(buf)
	var answer [][][]field.Element
	if err := dec.Decode(&answer); err != nil {
		return nil, err
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

// return true if the query inputs are invalid for IT schemes
func invalidQueryInputsIT(index, numServers int) bool {
	return index < 0 && numServers < 2
}

// return true if the query inputs are invalid for DPF-based schemes
func invalidQueryInputsDPF(index, numServers int) bool {
	return index < 0 && numServers != 2
}
