package client

import (
	"errors"
	"log"

	"github.com/cloudflare/circl/group"
	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/merkle"
	"github.com/si-co/vpir-code/lib/utils"
)

// Client represents the client for all (A)PIR clients implemented in the package
type Client interface {
	QueryBytes([]byte, int) ([][]byte, error)
	ReconstructBytes([][]byte) (interface{}, error)
}

// state of the client, used for all the schemes.
type state struct {
	// only used for Merkle tree-based approach and classic PIR
	ix int
	iy int

	// for multi-server
	alphas []uint32 // four alphas to meet desired soundness
	a      []uint32 // cointains [1, alpha_i], i = 0, .., 3

	// for single-server (DH)
	r  group.Scalar
	ht group.Element
}

// decodeAnswer decodes the gob-encoded answers from the servers and return
// them as slices of field elements.
func decodeAnswer(in [][]byte) ([][]uint32, error) {
	// decode all the answers one by one
	answer := make([][]uint32, len(in))
	for i, a := range in {
		answer[i] = utils.ByteSliceToUint32Slice(a)
	}

	return answer, nil
}

// reconstructPIR returns the database entry for the classical PIR schemes.
// These schemes are used as a baseline for the evaluation of the VPIR schemes.
func reconstructPIR(answers [][]byte, dbInfo *database.Info, state *state) ([]byte, error) {
	switch dbInfo.PIRType {
	case "classical", "":
		return reconstructValuePIR(answers, dbInfo, state)
	case "merkle":
		block, err := reconstructValuePIR(answers, dbInfo, state)
		if err != nil {
			return block, err
		}
		block = database.UnPadBlock(block)
		data := block[:len(block)-dbInfo.ProofLen]

		// check Merkle proof
		encodedProof := block[len(block)-dbInfo.ProofLen:]
		proof := merkle.DecodeProof(encodedProof)
		verified, err := merkle.VerifyProof(data, proof, dbInfo.Root)
		if err != nil {
			log.Fatalf("impossible to verify proof: %v", err)
		}
		if !verified {
			return nil, errors.New("REJECT!")
		}

		return data, nil
	default:
		panic("unknown PIRType")
	}
}

func reconstructValuePIR(answers [][]byte, dbInfo *database.Info, state *state) ([]byte, error) {
	// sum answers as vectors in GF(2)
	bs := dbInfo.BlockSize
	sum := make([]byte, bs)
	for k := range answers {
		fastxor.Bytes(sum, sum, answers[k][state.ix*bs:bs*(state.ix+1)])
	}

	return sum, nil
}

// return true if the query inputs are invalid for IT schemes
func invalidQueryInputsIT(index, numServers int) bool {
	return index < 0 && numServers < 2
}

func invalidQueryInputsFSS(numServers int) bool {
	return numServers != 2
}
