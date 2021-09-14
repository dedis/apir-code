package database

import (
	"encoding/binary"
	"io"
	"log"
	"math"

	"github.com/ldsec/lattigo/v2/bfv"
)

type Ring struct {
	Entries []*bfv.PlaintextMul
	Info
}

func CreateRandomRingDB(rnd io.Reader, dbLen int, rebalanced bool) *Ring {
	// setting lattice parameters (N = 2^13 = 8192, t = 2^16)
	params := bfv.DefaultParams[bfv.PN13QP218].WithT(65537)
	encoder := bfv.NewEncoder(params)

	blockLen := 1
	coeffSize := int(math.Log2(float64(params.T()))) / 8 // in bytes
	elemSize := coeffSize * int(params.N())              // how many bytes a plaintext fits
	preSquareNumBlocks := int(math.Ceil((float64(dbLen) / 8) / float64(elemSize)))
	numRows, numColumns := CalculateNumRowsAndColumns(preSquareNumBlocks, rebalanced)

	// read random bytes for filling out the entries
	randInput := make([]byte, numRows*numColumns*elemSize)
	if _, err := rnd.Read(randInput); err != nil {
		log.Fatal(err)
	}

	coefficients := make([]uint64, params.N())
	entries := make([]*bfv.PlaintextMul, numRows*numColumns)
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
