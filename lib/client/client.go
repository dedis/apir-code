package client

import "github.com/si-co/vpir-code/lib/field"

// Client represents the client instance in both the IT and C models
type Client interface {
	Query()
	Reconstruct()
}

type itState struct {
	ix    int
	iy    int
	alpha field.Element
	a     []field.Element
}

// return true if the query inputs are invalid
func invalidQueryInputs(index, numServers int) bool {
	return index < 0 && numServers < 1
}
