// Note: original file modified

// Copyright © 2018, 2019 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package merkle

import (
	"encoding/binary"
	"errors"
	"hash/adler32"
	"math"
)

// MerkleTree is the structure for the Merkle tree.
type MerkleTree struct {
	// hash is a pointer to the hashing struct
	hash HashType
	// data is the data from which the Merkle tree is created
	// data are stored as a map from the actual data encoded to string to
	// the index of the data in the tree
	data map[uint32]uint32
	// nodes are the leaf and branch nodes of the Merkle tree
	nodes [][]byte
}

func (t *MerkleTree) indexOf(input []byte) (uint32, error) {
	if i, ok := t.data[adler32.Checksum(input)]; ok {
		return i, nil
	}
	return 0, errors.New("data not found")
}

// GenerateProof generates the proof for a piece of data.
// If the data is not present in the tree this will return an error.
// If the data is present in the tree this will return the hashes for each level in the tree and the index of the value in the tree
func (t *MerkleTree) GenerateProof(data []byte) (*Proof, error) {
	// Find the index of the data
	index, err := t.indexOf(data)
	if err != nil {
		return nil, err
	}

	proofLen := int(math.Ceil(math.Log2(float64(len(t.data)))))
	hashes := make([][]byte, proofLen)

	cur := 0
	minI := uint32(math.Pow(2, float64(1))) - 1
	for i := index + uint32(len(t.nodes)/2); i > minI; i /= 2 {
		hashes[cur] = t.nodes[i^1]
		cur++
	}
	return newProof(hashes, index), nil
}

// EncodedProofLength returns the byte length of the proof for a piece of data.
// 4 bytes are for how many hashes are in the path, 8 bytes for embedding the index
// in the tree (see proof.go for details).
func (t *MerkleTree) EncodedProofLength() int {
	return int(math.Ceil(math.Log2(float64(len(t.data)))))*t.hash.HashLength() + numHashesByteSize + indexByteSize
}

// New creates a new Merkle tree using the provided raw data and default hash type.
// data must contain at least one element for it to be valid.
func New(data [][]byte) (*MerkleTree, error) {
	return NewUsing(data, NewBLAKE3())
}

// NewUsing creates a new Merkle tree using the provided raw data and supplied hash type.
// data must contain at least one element for it to be valid.
func NewUsing(data [][]byte, hash HashType) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, errors.New("tree must have at least 1 piece of data")
	}

	branchesLen := int(math.Exp2(math.Ceil(math.Log2(float64(len(data))))))

	// map with the original data to easily loop up the index
	md := make(map[uint32]uint32, len(data))
	// We pad our data length up to the power of 2
	nodes := make([][]byte, branchesLen+len(data)+(branchesLen-len(data)))
	// Leaves
	for i := range data {
		ib := indexToBytes(i)
		nodes[i+branchesLen] = hash.Hash(data[i], ib)
		md[adler32.Checksum(data[i])] = uint32(i)
	}
	for i := len(data) + branchesLen; i < len(nodes); i++ {
		nodes[i] = make([]byte, hash.HashLength())
	}

	// Branches
	for i := branchesLen - 1; i > 0; i-- {
		nodes[i] = hash.Hash(nodes[i*2], nodes[i*2+1])
	}

	tree := &MerkleTree{
		hash:  hash,
		nodes: nodes,
		data:  md,
	}

	return tree, nil
}

// Root returns the Merkle root (hash of the root node) of the tree.
func (t *MerkleTree) Root() []byte {
	return t.nodes[1]
}

// indexToBytes convert a data index in bytes representaiton
func indexToBytes(i int) []byte {
	if i > math.MaxUint32 {
		panic("index too big")
	}
	b := make([]byte, indexByteSize)
	binary.LittleEndian.PutUint32(b, uint32(i))
	return b
}
