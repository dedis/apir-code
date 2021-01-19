package client

import (
	"bytes"
	"encoding/gob"

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

// return true if the query inputs are invalid for IT schemes
func invalidQueryInputsIT(index, numServers int) bool {
	return index < 0 && numServers < 2
}

// return true if the query inputs are invalid for DPF-based schemes
func invalidQueryInputsDPF(index, numServers int) bool {
	return index < 0 && numServers != 2
}
