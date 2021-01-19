package client

import "github.com/si-co/vpir-code/lib/field"

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

// return true if the query inputs are invalid for IT schemes
func invalidQueryInputsIT(index, numServers int) bool {
	return index < 0 && numServers < 2
}

// return true if the query inputs are invalid for DPF-based schemes
func invalidQueryInputsDPF(index, numServers int) bool {
	return index < 0 && numServers != 2
}
