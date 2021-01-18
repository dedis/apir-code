package client

import (
	"errors"
	"github.com/si-co/vpir-code/lib/database"
	"io"
	"math/bits"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
)

// DPF represent the client for the DPF-based single- and multi-bit schemes
type DPF struct {
	rnd    io.Reader
	dbInfo database.Info
	state  *itState
}

func NewDPF(rnd io.Reader, info database.Info) *DPF {
	return &DPF{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

func (c *DPF) Query(index, numServers int) []dpf.DPFkey {
	if index < 0 {
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

	// Compute the position in the db (vector or matrix)
	ix := index % c.dbInfo.NumColumns
	// if db is a vector, iy always equals 0
	iy := index / c.dbInfo.NumColumns
	// set ITClient state
	// set state
	c.state = &itState{
		ix:    ix,
		iy:    iy,
		alpha: *alpha,
		a:     a[1:],
	}

	// client initialization is the same for both single- and multi-bit scheme
	key0, key1 := dpf.Gen(uint64(ix), a, uint64(bits.Len(uint(c.dbInfo.NumColumns))))

	return []dpf.DPFkey{key0, key1}
}

func (c *DPF) Reconstruct(answers [][][]field.Element) ([]field.Element, error) {
	sum := make([][]field.Element, c.dbInfo.NumRows)

	if c.dbInfo.BlockSize == cst.SingleBitBlockLength {
		// sum answers as vectors in F^b
		for i := 0; i < c.dbInfo.NumRows; i++ {
			sum[i] = make([]field.Element, 1)
			for k := range answers {
				sum[i][0].Add(&sum[i][0], &answers[k][i][0])
			}
		}
		for i := 0; i < c.dbInfo.NumRows; i++ {
			if i == c.state.iy {
				switch {
				case sum[i][0].Equal(&c.state.alpha):
					return []field.Element{cst.One}, nil
				case sum[i][0].Equal(&cst.Zero):
					return []field.Element{cst.Zero}, nil
				default:
					return nil, errors.New("REJECT!")
				}
			} else {
				if !sum[i][0].Equal(&c.state.alpha) && !sum[i][0].Equal(&cst.Zero) {
					return nil, errors.New("REJECT!")
				}
			}
		}
	}

	// sum answers as vectors in F^(b+1)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		sum[i] = make([]field.Element, c.dbInfo.BlockSize+1)
		for b := 0; b < c.dbInfo.BlockSize+1; b++ {
			for k := range answers {
				sum[i][b].Add(&sum[i][b], &answers[k][i][b])
			}
		}
	}
	var tag, prod field.Element
	messages := make([]field.Element, c.dbInfo.BlockSize)
	for i := 0; i < c.dbInfo.NumRows; i++ {
		copy(messages, sum[i][:len(sum[i])-1])
		tag = sum[i][len(sum[i])-1]
		// compute reconstructed tag
		reconstructedTag := field.Zero()
		for b := 0; b < len(messages); b++ {
			prod.Mul(&c.state.a[b], &messages[b])
			reconstructedTag.Add(&reconstructedTag, &prod)
		}
		if !tag.Equal(&reconstructedTag) {
			return nil, errors.New("REJECT")
		}
	}

	return sum[c.state.iy][:len(sum[c.state.iy])-1], nil
}