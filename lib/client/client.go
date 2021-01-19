package client

import (
	"bytes"
	"encoding/gob"

	"github.com/si-co/vpir-code/lib/field"
)

// Client represents the client instance in both the IT and DPF-based schemes
type Client interface {
	QueryBytes([]byte) ([]byte, error)
	ReconstructBytes([]byte) ([]byte, error)
}

type state struct {
	ix    int
	iy    int
	alpha field.Element
	a     []field.Element
}

type queryInputs struct {
	index      int
	numServers int
}

// general functions for both IT and DPF-based clients
func reconstrucBytes(a []byte) ([]byte, error) {
	// decode answer
	buf := bytes.NewBuffer(a)
	dec := gob.NewDecoder(buf)
	var answer [][][]field.Element
	if err := dec.Decode(&answer); err != nil {
		return nil, err
	}

	// get reconstruction
	r, err := c.Reconstruct(answer)
	if err != nil {
		return nil, err
	}

	// encode reconstruction
	buf.Reset()
	enc := gob.NewEncoder(buf)
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
