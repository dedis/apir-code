package client

import (
	"github.com/si-co/vpir-code/lib/database"
	"io"
	"math/bits"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

// DPF represent the client for the DPF-based single- and multi-bit schemes
type DPF struct {
	rnd    io.Reader
	dbInfo database.Info
	state  *dpfState
}

type dpfState struct {
	i     int
	alpha field.Element
	a     []field.Element
}

func NewDPF(rnd io.Reader, info database.Info) *DPF {
	return &DPF{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

func (c *DPF) Query(index, numServers int) []dpf.DPFkey {
	if index < 0 || index > cst.DBLength {
		panic("query index out of bound")
	}
	if numServers < 1 {
		panic("need at least 1 server")
	}
	if numServers != 2 {
		panic("DPF implementation only works with 2 servers")
	}

	// sample random alpha
	alpha, err := new(field.Element).SetRandom(c.rnd)
	if err != nil {
		panic(err)
	}

	var a []field.Element
	if c.dbInfo.BlockSize != cst.SingleBitBlockLength {
		a = field.PowerVectorWithOne(*alpha, c.dbInfo.BlockSize)
	} else {
		// the single-bit scheme needs a single alpha
		a = make([]field.Element, 1)
		a[0] = *alpha
	}

	// set ITClient state
	c.state = &dpfState{i: index, alpha: *alpha, a: a[1:]}

	// client initialization is the same for both single- and multi-bit scheme
	key0, key1 := dpf.Gen(uint64(index), a, uint64(bits.Len(uint(constants.DBLength))))

	return []dpf.DPFkey{key0, key1}
}

func (c *DPF) Reconstruct(answers [][]field.Element) ([]field.Element, error) {
	index := 0

	return reconstruct(answers, blockSize, index, c.state.alpha, c.state.a)
}
