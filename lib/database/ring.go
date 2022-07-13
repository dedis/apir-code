package database

import (
	"encoding/binary"
	"io"
	"log"
	"math"

	"github.com/tuneinsight/lattigo/v3/bfv"
)

type Ring struct {
	Entries []*bfv.PlaintextMul
	Info
}

func CreateRandomRingDB(rnd io.Reader, dbLen int, rebalanced bool) *Ring {
	// setting lattice parameters (N = 2^13 = 8192, t = 2^16)
	// PN13QP218 is a set of default parameters with logN=13 and logQP=218
	// PN13QP218 = ParametersLiteral{
	// 	LogN:  13,
	// 	T:     65537,
	// 	Q:     []uint64{0x3fffffffef8001, 0x4000000011c001, 0x40000000120001},
	// 	P:     []uint64{0x7ffffffffb4001},
	// 	Sigma: rlwe.DefaultSigma,
	// }
	paramsDef := bfv.PN13QP218
	paramsDef.T = 65537
	params, err := bfv.NewParametersFromLiteral(paramsDef)
	if err != nil {
		log.Fatal(err)
	}
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
		encoder.EncodeMul(coefficients, entries[i])
	}

	return &Ring{Entries: entries,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
			LatParams: params,
		},
	}
}
