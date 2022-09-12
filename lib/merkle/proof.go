// Copyright © 2018, 2019 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package merkle

import (
	"bytes"
	"encoding/binary"
)

const (
	indexByteSize     = 4
	numHashesByteSize = 4
)

// Proof is a proof of a Merkle tree
type Proof struct {
	Hashes [][]byte
	Index  uint32
}

// newProof generates a Merkle proof
func newProof(hashes [][]byte, index uint32) *Proof {
	return &Proof{
		Hashes: hashes,
		Index:  index,
	}
}

// VerifyProof verifies a Merkle tree proof for a piece of data using the default hash type.
// The proof and path are as per Merkle tree's GenerateProof(), and root is the root hash of the tree against which the proof is to
// be verified.  Note that this does not require the Merkle tree to verify the proof, only its root; this allows for checking
// against historical trees without having to instantiate them.
//
// This returns true if the proof is verified, otherwise false.
func VerifyProof(data []byte, proof *Proof, root []byte) (bool, error) {
	return VerifyProofUsing(data, proof, root, NewBLAKE3())
}

// VerifyProofUsing verifies a Merkle tree proof for a piece of data using the provided hash type.
// The proof and is as per Merkle tree's GenerateProof(), and root is the root hash of the tree against which the proof is to
// be verified.  Note that this does not require the Merkle tree to verify the proof, only its root; this allows for checking
// against historical trees without having to instantiate them.
//
// This returns true if the proof is verified, otherwise false.
func VerifyProofUsing(data []byte, proof *Proof, root []byte, hashType HashType) (bool, error) {
	proofHash := generateProofHash(data, proof, hashType)
	if bytes.Equal(root, proofHash) {
		return true, nil
	}
	return false, nil
}

func generateProofHash(data []byte, proof *Proof, hashType HashType) []byte {
	var proofHash []byte
	ib := indexToBytes(int(proof.Index))
	proofHash = hashType.Hash(data, ib)
	index := proof.Index + (1 << uint(len(proof.Hashes)))

	for _, hash := range proof.Hashes {
		if index%2 == 0 {
			proofHash = hashType.Hash(proofHash, hash)
		} else {
			proofHash = hashType.Hash(hash, proofHash)
		}
		index = index >> 1
	}
	return proofHash
}

func DecodeProof(p []byte) *Proof {
	// number of hashes
	numHashes := binary.LittleEndian.Uint32(p[:numHashesByteSize])

	// hashes
	hashLength := uint32(32) // blake3
	hashes := make([][]byte, numHashes)
	for i := uint32(0); i < numHashes; i++ {
		hashes[i] = p[4+hashLength*i : 4+hashLength*(i+1)]
	}

	// index
	index := binary.LittleEndian.Uint32(p[len(p)-indexByteSize:])

	return &Proof{
		Hashes: hashes,
		Index:  index,
	}
}

func EncodeProof(p *Proof) []byte {
	// out length is 4 bytes for numHashes, number of bytes for the hashes
	// (32 bytes for each hash) and 8 bytes for encoded index
	outLen := numHashesByteSize + len(p.Hashes)*32 + indexByteSize
	out := make([]byte, outLen)

	// encode number of hashes
	numHashes := uint32(len(p.Hashes))
	b := make([]byte, numHashesByteSize)
	binary.LittleEndian.PutUint32(b, numHashes)
	//out = append(out, b...)
	copy(out[:numHashesByteSize], b)

	// encode hashes
	for i, h := range p.Hashes {
		//out = append(out, h...)
		copy(out[numHashesByteSize+i*len(h):numHashesByteSize+(i+1)*len(h)], h)
	}

	// encode index
	b1 := make([]byte, indexByteSize)
	binary.LittleEndian.PutUint32(b1, p.Index)
	copy(out[len(out)-indexByteSize:], b1)
	//out = append(out, b1...)

	return out
}
