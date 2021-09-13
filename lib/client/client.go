package client

import (
	"errors"
	"log"

	"github.com/cloudflare/circl/group"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/merkle"
	"github.com/si-co/vpir-code/lib/utils"
)

// Client represents the client instance in both the IT and DPF-based schemes.
type Client interface {
	// TODO: modify interface to take []byte instead of index for the first
	// parameter, so that we can encode more complex queries here
	QueryBytes(int, int) ([][]byte, error)
	ReconstructBytes([][]byte) (interface{}, error)
}

// state of the client, used for all the schemes.
type state struct {
	// only used for Merkle tree-based approach and classic PIR
	ix int
	iy int

	// for multi-server
	alphas []uint32 // four alphas to meet desired soundness
	a      []uint32 // cointains the four [1, alpha_i] sub-vectors

	// for single-server (DH)
	r   group.Scalar
	ht  group.Element
	key *bfv.SecretKey // lattice secret key
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

/* // generateClientState returns the client state with all the needed settings.
func generateClientState(index int, rnd io.Reader, dbInfo *database.Info) (*state, error) {
	// initialize state
	st := &state{}

	// sample random alpha
	st.alpha = field.RandElementWithPRG(rnd)

	// Compute the position in the db (vector or matrix)
	// if db is a vector, ix always equals 0
	st.ix = index / dbInfo.NumColumns
	st.iy = index % dbInfo.NumColumns

	// compute vector a = (1, alpha, alpha^2, ..., alpha^b) for the
	// multi-bit scheme
	// +1 to BlockSize for recovering true value
	st.a = make([]uint32, dbInfo.BlockSize+1)
	st.a[0] = 1
	for i := 1; i < len(st.a); i++ {
		a := (uint64(st.a[i-1]) * uint64(st.alpha)) % uint64(field.ModP)
		st.a[i] = uint32(a)
	}

	return st, nil
} */

// reconstruct takes as input the answers fro mthe servers, the info about the
// database and the client state to return the reconstructed database entry.
// The integrity check is performed in this function.
func reconstruct(answers [][]uint32, dbInfo *database.Info, st *state) ([]uint32, error) {
	sum := make([][]uint32, dbInfo.NumRows)

	// mutli-bit scheme
	// sum answers as vectors in F^(b+1)
	for i := 0; i < dbInfo.NumRows; i++ {
		sum[i] = make([]uint32, dbInfo.BlockSize+1)
		for b := 0; b < dbInfo.BlockSize+1; b++ {
			for k := range answers {
				sum[i][b] = (sum[i][b] + answers[k][i*(dbInfo.BlockSize+1)+b]) % field.ModP
			}
		}
	}

	messages := make([]uint32, dbInfo.BlockSize)
	for i := 0; i < dbInfo.NumRows; i++ {
		copy(messages, sum[i][:len(sum[i])-1])
		tag := sum[i][len(sum[i])-1]
		// compute reconstructed tag
		reconstructedTag := uint32(0)
		for b := 0; b < len(messages); b++ {
			p := (uint64(st.a[b+1]) * uint64(messages[b])) % uint64(field.ModP)
			reconstructedTag = (reconstructedTag + uint32(p)) % field.ModP
		}

		if tag != reconstructedTag {
			return nil, errors.New("REJECT")
		}
	}

	//return sum[st.ix][:len(sum[st.ix])-1], nil
	return sum[0][:len(sum[0])-1], nil
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
	sum := make([][]byte, dbInfo.NumRows)

	// sum answers as vectors in GF(2)
	bs := dbInfo.BlockSize
	for i := 0; i < dbInfo.NumRows; i++ {
		sum[i] = make([]byte, dbInfo.BlockSize)
		for k := range answers {
			fastxor.Bytes(sum[i], sum[i], answers[k][i*bs:bs*(i+1)])
		}
	}

	return sum[state.ix], nil
}

// return true if the query inputs are invalid for IT schemes
func invalidQueryInputsIT(index, numServers int) bool {
	return index < 0 && numServers < 2
}

// return true if the query inputs are invalid for DPF-based schemes
func invalidQueryInputsDPF(index, numServers int) bool {
	return index < 0 && numServers != 2
}

func invalidQueryInputsFSS(numServers int) bool {
	return numServers != 2
}
