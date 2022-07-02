package client

import "github.com/si-co/vpir-code/lib/database"

// LEW based authenticated single server PIR client

// Client description
type LWE struct {
	dbInfo *database.Info
}

func NewLWE() *LWE {
}

func (c *LWE) QueryBytes(index int) ([]byte, error) {

}

func (c *LWE) ReconstructBytes(a []byte) ([]uint64, error) {

}
