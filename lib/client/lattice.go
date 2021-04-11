// This code is partially based on the example from
// https://github.com/ldsec/lattigo/blob/master/examples/dbfv/pir/main.go
package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/si-co/vpir-code/lib/monitor"
	"math"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/si-co/vpir-code/lib/database"
)

type Lattice struct {
	dbInfo *database.Info
	state  *state
}

type EncodedQuery struct {
	Ciphertext   []byte
	RotationKeys []byte
}

// NewLattice returns a client for the lattice-based single-server multi-bit
// scheme, working both with the vector and the matrix representation of
// the database.
func NewLattice(info *database.Info) *Lattice {
	return &Lattice{
		dbInfo: info,
		state:  nil,
	}
}

func (c *Lattice) QueryBytes(index int) ([]byte, error) {
	params := c.dbInfo.LatParams
	encoder := bfv.NewEncoder(params)

	timer := monitor.NewMonitor()
	// Key generation
	kgen := bfv.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPair()
	rtk := kgen.GenRotationKeysPow2(sk)
	encryptor := bfv.NewEncryptorFromPk(params, pk)
	// saving to the client state
	c.state = &state{
		ix:  index / c.dbInfo.NumColumns,
		iy:  index % c.dbInfo.NumColumns,
		key: sk,
	}
	fmt.Printf("Time to gen keys: %v\n", timer.RecordAndReset())

	encQuery := genQuery(params, c.state.iy, encoder, encryptor)
	fmt.Printf("Time to gen query: %v\n", timer.RecordAndReset())

	encodedQuery, err := encodeQuery(encQuery, rtk)
	fmt.Printf("Time to encode query: %v\n", timer.RecordAndReset())
	if err != nil {
		return nil, err
	}

	return encodedQuery, nil
}

func (c *Lattice) ReconstructBytes(a []byte) ([][]byte, error) {
	var err error
	params := c.dbInfo.LatParams
	encoder := bfv.NewEncoder(params)
	decryptor := bfv.NewDecryptor(params, c.state.key)
	ciphertextSize := len(a) / c.dbInfo.NumRows

	ctx := new(bfv.Ciphertext)
	var coeffs []uint64
	tmp := make([]byte, 8)
	column := make([][]byte, c.dbInfo.NumRows)
	dataSize := int(math.Log2(float64(params.T()))) / 8
	j := 0
	for i := 0; i < len(a); i += ciphertextSize {
		err = ctx.UnmarshalBinary(a[i : i+ciphertextSize])
		if err != nil {
			return nil, err
		}
		coeffs = encoder.DecodeUintNew(decryptor.DecryptNew(ctx))
		column[j] = make([]byte, 0, dataSize*int(params.N()))
		for _, coeff := range coeffs {
			binary.BigEndian.PutUint64(tmp, coeff)
			column[j] = append(column[j], tmp[len(tmp)-dataSize:]...)
		}
		j++
		//if i / ciphertextSize == c.state.ix {
		//	return m, nil
		//}
	}

	return column, nil
}

func genQuery(params *bfv.Parameters, queryIndex int, encoder bfv.Encoder, encryptor bfv.Encryptor) *bfv.Ciphertext {
	// Query ciphertext
	queryCoeffs := make([]uint64, params.N())
	queryCoeffs[queryIndex] = 1
	query := bfv.NewPlaintext(params)

	var encQuery *bfv.Ciphertext
	encoder.EncodeUint(queryCoeffs, query)
	encQuery = encryptor.EncryptNew(query)

	return encQuery
}

func encodeQuery(ct *bfv.Ciphertext, rtk *bfv.RotationKeys) ([]byte, error) {
	// Marshal all the keys
	ect, err := ct.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ertk, err := rtk.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Encode the keys as struct with gob
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&EncodedQuery{
		Ciphertext:   ect,
		RotationKeys: ertk,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
