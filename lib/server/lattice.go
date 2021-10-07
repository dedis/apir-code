// This code is partially based on the example from
// https://github.com/ldsec/lattigo/blob/master/examples/dbfv/pir/main.go
package server

import (
	"bytes"
	"encoding/gob"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"log"
	"runtime"
)

// Lattice is the server for the computational multi-bit scheme
type Lattice struct {
	db *database.Ring
}

// NewLattice return a server for a lattice-based single-server scheme,
// working both with the vector and the rebalanced representation of the
// database.
func NewLattice(db *database.Ring) *Lattice {
	return &Lattice{db: db}
}

func (s *Lattice) DBInfo() *database.Info {
	return &s.db.Info
}

// AnswerBytes decodes the input, computes reply and encodes the output
func (s *Lattice) AnswerBytes(q []byte) ([]byte, error) {
	params := s.db.LatParams
	encoder := bfv.NewEncoder(params)

	// get ciphertext, rotation keys and relinearization key from the query
	ctx, rtk, err := s.unmarshalQuery(q)
	if err != nil {
		log.Fatal(err)
	}

	// compute masks to expand the compressed ciphertext to NumColumns ciphertexts
	plainMask := make([]*bfv.PlaintextMul, s.db.NumColumns)
	// Plaintext masks: plainmask[i] = encode([0, ..., 0, 1_i, 0, ..., 0])
	// (zero with a 1 at the i-th position).
	for i := range plainMask {
		maskCoeffs := make([]uint64, params.N())
		maskCoeffs[i] = 1
		plainMask[i] = bfv.NewPlaintextMul(params)
		encoder.EncodeUintMul(maskCoeffs, plainMask[i])
	}

	// multithreading
	NGoRoutines := runtime.NumCPU()
	// make sure that we do not need up with routines processing 0 elements
	if NGoRoutines > s.db.NumRows {
		NGoRoutines = s.db.NumRows
	}
	rowsPerRoutine := s.db.NumRows / NGoRoutines
	replies := make([]chan []*bfv.Ciphertext, NGoRoutines)
	var begin, end int
	for i := 0; i < NGoRoutines; i++ {
		begin, end = i*rowsPerRoutine, (i+1)*rowsPerRoutine
		// make the last routine take all the left-over (from division) rows
		if i == NGoRoutines-1 {
			end = s.db.NumRows
		}
		replyChan := make(chan []*bfv.Ciphertext, 1)
		replies[i] = replyChan
		go s.processRows(begin, end, ctx, plainMask, rtk, replyChan)
	}

	answer := make([]*bfv.Ciphertext, 0, s.db.NumRows)
	for i, reply := range replies {
		chunk := <-reply
		answer = append(answer, chunk...)
		close(replies[i])
	}

	encoded, err := s.marshalAnswer(answer)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func (s *Lattice) processRows(begin, end int, encQuery *bfv.Ciphertext, masks []*bfv.PlaintextMul,
	rtk *bfv.RotationKeys, replyTo chan<- []*bfv.Ciphertext) {
	replies := make([]*bfv.Ciphertext, end-begin)
	evaluator := bfv.NewEvaluator(s.db.LatParams)
	for i := begin; i < end; i++ {
		prod := bfv.NewCiphertext(s.db.LatParams, 1)
		for j := 0; j < s.db.NumColumns; j++ {
			tmp := bfv.NewCiphertext(s.db.LatParams, 1)
			// 1) Multiplication of the query with the plaintext mask
			evaluator.Mul(encQuery, masks[j], tmp)
			// 2) Inner sum (populate all the slots with the sum of all the slots)
			evaluator.InnerSum(tmp, rtk, tmp)
			// 3) Multiplication of 2) with the (i,j)-th plaintext of the db
			evaluator.Mul(tmp, s.db.Entries[i*s.db.NumColumns+j], tmp)
			// 4) Add the result of the column multiplication to the final row product
			evaluator.Add(prod, tmp, prod)
		}
		// save the row product
		replies[i-begin] = prod
	}
	replyTo <- replies
}

func (s *Lattice) unmarshalQuery(query []byte) (*bfv.Ciphertext, *bfv.RotationKeys, error) {
	var err error
	// decode query
	buf := bytes.NewBuffer(query)
	dec := gob.NewDecoder(buf)
	var decoded client.EncodedQuery
	if err = dec.Decode(&decoded); err != nil {
		return nil, nil, err
	}

	ctx := new(bfv.Ciphertext)
	err = ctx.UnmarshalBinary(decoded.Ciphertext)
	if err != nil {
		return nil, nil, err
	}
	rtk := new(bfv.RotationKeys)
	err = rtk.UnmarshalBinary(decoded.RotationKeys)
	if err != nil {
		return nil, nil, err
	}

	return ctx, rtk, nil
}

func (s *Lattice) marshalAnswer(answer []*bfv.Ciphertext) ([]byte, error) {
	example, err := answer[0].MarshalBinary()
	if err != nil {
		return nil, err
	}
	encoded := make([]byte, 0, len(answer)*len(example))
	for _, a := range answer {
		e, err := a.MarshalBinary()
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, e...)
	}
	return encoded, nil
}
