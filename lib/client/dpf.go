package client

import (
	"errors"
	"fmt"
	"io"
	"math/bits"

	"github.com/si-co/vpir-code/lib/constants"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

// DPF represent the client for the DPF-based single- and multi-bit schemes
type DPF struct {
	rnd   io.Reader
	state *dpfState
}

type dpfState struct {
	i     int
	alpha field.Element
	a     []field.Element
}

func NewDPF(rnd io.Reader) *DPF {
	return &DPF{
		rnd:   rnd,
		state: nil,
	}
}

func (c *DPF) Query(index, blockSize, numServers int) ([][]byte, [][]dpf.FssKeyEq2P) {
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
	if blockSize != cst.SingleBitBlockLength {
		a = field.PowerVectorWithOne(*alpha, blockSize)
	} else {
		fmt.Println("here")
		// the single-bit scheme needs a single alpha
		a = make([]field.Element, 1)
		a[0] = *alpha
	}

	// set ITClient state
	c.state = &dpfState{i: index, alpha: *alpha, a: a[1:]}

	// client initialization is the same for both single- and multi-bit scheme
	fClient := dpf.ClientInitialize(uint(bits.Len(uint(constants.DBLength))))

	// compute dpf keys
	fssKeysVector := make([][]dpf.FssKeyEq2P, 2)
	if blockSize != cst.SingleBitBlockLength {
		fssKeysVector = fClient.GenerateTreePFVector(uint(index), alpha, blockSize)
	} else {
		fssKeys := fClient.GenerateTreePF(uint(index), alpha)
		fssKeysVector[0] = append(fssKeysVector[0], fssKeys[0])
		fssKeysVector[1] = append(fssKeysVector[1], fssKeys[1])
	}

	return fClient.PrfKeys, fssKeysVector
}

func (c *DPF) Reconstruct(answers [][]field.Element, blockSize int) ([]field.Element, error) {
	answersLen := len(answers[0])
	sum := make([]field.Element, answersLen)

	// sum answers as vectors in F(2^128)^(1+b)
	for i := 0; i < answersLen; i++ {
		sum[i] = field.Zero()
		for s := range answers {
			sum[i].Add(&sum[i], &answers[s][i])
		}
	}

	// select index depending on the matrix representation
	if blockSize == cst.SingleBitBlockLength {
		switch {
		case sum[0].Equal(&c.state.alpha):
			return []field.Element{cst.One}, nil
		case sum[0].Equal(&cst.Zero):
			return []field.Element{cst.Zero}, nil
		default:
			return nil, errors.New("REJECT!")
		}
	}

	tag := sum[len(sum)-1]
	messages := sum[:len(sum)-1]

	// compute reconstructed tag
	reconstructedTag := field.Zero()
	for i := 0; i < len(messages); i++ {
		var prod field.Element
		prod.Mul(&c.state.a[i], &messages[i])
		reconstructedTag.Add(&reconstructedTag, &prod)
	}

	if !tag.Equal(&reconstructedTag) {
		return nil, errors.New("REJECT")
	}

	return messages, nil
}
