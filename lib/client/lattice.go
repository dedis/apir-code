// This code is partially based on the example from
// https://github.com/ldsec/lattigo/blob/master/examples/dbfv/pir/main.go
package client

import (
	"bytes"
	"encoding/gob"
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

func (c *Lattice) QueryBytes(index int, db *database.Ring) ([]byte, error) {
	params := c.dbInfo.LatParams
	encoder := bfv.NewEncoder(params)

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

	encQuery := genQuery(params, index, encoder, encryptor)

	//// Debugging
	//plainMask := make([]*bfv.PlaintextMul, c.dbInfo.NumColumns)
	//// Plaintext masks: plainmask[i] = encode([0, ..., 0, 1_i, 0, ..., 0])
	//// (zero with a 1 at the i-th position).
	//for i := range plainMask {
	//	maskCoeffs := make([]uint64, params.N())
	//	maskCoeffs[i] = 1
	//	plainMask[i] = bfv.NewPlaintextMul(params)
	//	encoder.EncodeUintMul(maskCoeffs, plainMask[i])
	//}
	//
	//decryptor := bfv.NewDecryptor(params, c.state.key)
	//evaluator := bfv.NewEvaluator(params)
	//for j := 0; j < c.dbInfo.NumColumns; j++ {
	//	//for j := 0; j < 1; j++ {
	//	tmp := bfv.NewCiphertext(params, 1)
	//	// 1) Multiplication of the query with the plaintext mask
	//	evaluator.Mul(encQuery, plainMask[j], tmp)
	//	// 2) Inner sum (populate all the slots with the sum of all the slots)
	//	evaluator.InnerSum(tmp, rtk, tmp)
	//	fmt.Printf("%d ", encoder.DecodeUintNew(decryptor.DecryptNew(tmp))[0])
	//	// 3) Multiplication of 2) with the (i,j)-th plaintext of the db
	//	fmt.Printf("%d ", encoder.DecodeUintNew(db.Entries[j])[0])
	//	evaluator.Mul(tmp, db.Entries[j], tmp)
	//	fmt.Println(encoder.DecodeUintNew(decryptor.DecryptNew(tmp))[0])
	//	// 4) Add the result of the column multiplication to the final row product
	//	//evaluator.Add(prodDeg2, tmp, prodDeg2)
	//}
	////fmt.Println(encoder.DecodeUintNew(decryptor.DecryptNew(evaluator.RelinearizeNew(prodDeg2, rlt))))
	//
	////

	encodedQuery, err := encodeQuery(encQuery, rtk)
	if err != nil {
		return nil, err
	}

	return encodedQuery, nil
}

func (c *Lattice) ReconstructBytes(a []byte) ([]uint64, error) {
	var err error
	params := c.dbInfo.LatParams
	encoder := bfv.NewEncoder(params)
	decryptor := bfv.NewDecryptor(params, c.state.key)
	ciphertextSize := len(a)/c.dbInfo.NumRows

	ctx := new(bfv.Ciphertext)
	var m []uint64
	for i := 0; i < len(a); i += ciphertextSize {
		err = ctx.UnmarshalBinary(a[i:i+ciphertextSize])
		if err != nil {
			return nil, err
		}
		m = encoder.DecodeUintNew(decryptor.DecryptNew(ctx))
		if i / ciphertextSize == c.state.ix {
			return m, nil
		}
	}

	return nil, nil
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
