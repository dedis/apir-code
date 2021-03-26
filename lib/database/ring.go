package database

import (
	"encoding/binary"
	"github.com/ldsec/lattigo/v2/bfv"
	helpers "github.com/si-co/vpir-code/lib/utils"
	"io"
	"log"
	"math"
)

type Ring struct {
	Entries []*bfv.PlaintextMul
	Info
}

func CreateRandomRingDB(rnd io.Reader, dbLen int, rebalanced bool) *Ring {
	// setting lattice parameters (N = 2^13 = 8192, t = 2^16)
	params := bfv.DefaultParams[bfv.PN13QP218].WithT(65537)
	encoder := bfv.NewEncoder(params)

	coeffSize := int(math.Log2(float64(params.T()))) / 8
	blockLen := coeffSize * int(params.N())
	numBlocks := dbLen / (8 * blockLen)
	// handle very small db
	if numBlocks == 0 {
		numBlocks = 1
	}
	var numColumns, numRows int
	if rebalanced {
		helpers.IncreaseToNextSquare(&numBlocks)
		numColumns = int(math.Sqrt(float64(numBlocks)))
		numRows = numColumns
	} else {
		numColumns = numBlocks
		numRows = 1
	}

	// read random bytes for filling out the entries
	randInput := make([]byte, numBlocks*blockLen)
	if _, err := rnd.Read(randInput); err != nil {
		log.Fatal(err)
	}

	coefficients := make([]uint64, params.N())
	entries := make([]*bfv.PlaintextMul, numBlocks)
	tmp := make([]byte, 8)
	for i := range entries {
		entries[i] = bfv.NewPlaintextMul(params)
		for l := range coefficients {
			copy(tmp[len(tmp)-2:], randInput[(i*int(params.N())+l)*coeffSize:(i*int(params.N())+l+1)*coeffSize])
			coefficients[l] = binary.BigEndian.Uint64(tmp)
		}
		encoder.EncodeUintMul(coefficients, entries[i])
	}

	return &Ring{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
			LatParams: params,
		},
	}
}
