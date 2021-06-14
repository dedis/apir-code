package client

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/cloudflare/circl/group"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/lukechampine/fastxor"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/merkle"
)

// Client represents the client instance in both the IT and DPF-based schemes.
type Client interface {
	QueryBytes(int, int) ([][]byte, error)
	ReconstructBytes([][]byte) (interface{}, error)
}

// state of the client, used for all the schemes.
type state struct {
	ix int
	iy int

	// for multi-server
	alpha field.Element
	a     []field.Element

	// for single-server (DH)
	r   group.Scalar
	ht  group.Element
	key *bfv.SecretKey // lattice secret key
}

// decodeAnswer decodes the gob-encoded answers from the servers and return
// them as slices of field elements.
func decodeAnswer(a [][]byte) ([][]field.Element, error) {
	// decode all the answers one by one
	answer := make([][]field.Element, len(a))
	for i, ans := range a {
		buf := bytes.NewBuffer(ans)
		dec := gob.NewDecoder(buf)
		var serverAnswer []field.Element
		if err := dec.Decode(&serverAnswer); err != nil {
			return nil, err
		}
		answer[i] = serverAnswer
	}

	return answer, nil
}

// generateClientState returns the client state with all the needed settings.
func generateClientState(index int, rnd io.Reader, dbInfo *database.Info) (*state, error) {
	// initialize state
	st := &state{}

	// sample random alpha
	if _, err := st.alpha.SetRandom(rnd); err != nil {
		return nil, err
	}

	// Compute the position in the db (vector or matrix)
	// if db is a vector, ix always equals 0
	st.ix = index / dbInfo.NumColumns
	st.iy = index % dbInfo.NumColumns

	if dbInfo.BlockSize != cst.SingleBitBlockLength {
		// compute vector a = (1, alpha, alpha^2, ..., alpha^b) for the
		// multi-bit scheme
		// +1 to BlockSize for recovering true value
		st.a = make([]field.Element, dbInfo.BlockSize+1)
		st.a[0] = field.One()
		for i := 1; i < len(st.a); i++ {
			st.a[i].Mul(&st.a[i-1], &st.alpha)
		}
	} else {
		// the single-bit scheme needs a single alpha
		st.a = make([]field.Element, 1)
		st.a[0] = st.alpha
	}

	return st, nil
}

// reconstruct takes as input the answers fro mthe servers, the info about the
// database and the client state to return the reconstructed database entry.
// The integrity check is performed in this function.
func reconstruct(answers [][]field.Element, dbInfo *database.Info, st *state) ([]field.Element, error) {
	sum := make([][]field.Element, dbInfo.NumRows)

	// single-bit scheme
	if dbInfo.BlockSize == cst.SingleBitBlockLength {
		// sum answers as vectors in F^b
		for i := 0; i < dbInfo.NumRows; i++ {
			sum[i] = make([]field.Element, 1)
			for k := range answers {
				sum[i][0].Add(&sum[i][0], &answers[k][i])
			}
		}
		// verify integrity and return database entry if accept
		for i := 0; i < dbInfo.NumRows; i++ {
			if i == st.ix {
				switch {
				case sum[i][0].Equal(&st.alpha):
					return []field.Element{cst.One}, nil
				case sum[i][0].Equal(&cst.Zero):
					return []field.Element{cst.Zero}, nil
				default:
					return nil, errors.New("REJECT!")
				}
			} else {
				if !sum[i][0].Equal(&st.alpha) && !sum[i][0].Equal(&cst.Zero) {
					return nil, errors.New("REJECT!")
				}
			}
		}
	}

	// mutli-bit scheme
	// sum answers as vectors in F^(b+1)
	for i := 0; i < dbInfo.NumRows; i++ {
		sum[i] = make([]field.Element, dbInfo.BlockSize+1)
		for b := 0; b < dbInfo.BlockSize+1; b++ {
			for k := range answers {
				sum[i][b].Add(&sum[i][b], &answers[k][i*(dbInfo.BlockSize+1)+b])
			}
		}
	}

	var prod field.Element
	messages := make([]field.Element, dbInfo.BlockSize)
	for i := 0; i < dbInfo.NumRows; i++ {
		copy(messages, sum[i][:len(sum[i])-1])
		tag := sum[i][len(sum[i])-1]
		// compute reconstructed tag
		reconstructedTag := field.Zero()
		for b := 0; b < len(messages); b++ {
			prod.Mul(&st.a[b+1], &messages[b])
			reconstructedTag.Add(&reconstructedTag, &prod)
		}
		if !tag.Equal(&reconstructedTag) {
			fmt.Println(tag)
			fmt.Println(reconstructedTag)
			return nil, errors.New("REJECT")
		}
	}

	return sum[st.ix][:len(sum[st.ix])-1], nil
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
		data := block[:len(block)-dbInfo.ProofLen]

		// check Merkle proof
		encodedProof := block[dbInfo.BlockSize-dbInfo.ProofLen:]
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
