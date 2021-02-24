package database

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/field"
	merkletree "github.com/wealdtech/go-merkletree"
)

type Merkle struct {
	Entries [][]byte
	Info

	Root        []byte
	ProofLength int
}

// CreateRandomMultiBitMerkle
// blockLen is the number of byte in a block, as byte is viewd as an element in this
// case
func CreateRandomMultiBitMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Merkle {
	db := CreateRandomMultiBitBytes(rnd, dbLen, numRows, blockLen)

	// a leaf contains a block, which has the same numbers of bytes as a
	// block composed of field elements
	blockLenBytes := blockLen * field.Bytes
	blocks := entriesToBlocks(db.Entries, blockLenBytes)
	tree, err := merkletree.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}

	// get the root hash of the tree
	root := tree.Root()

	// generate and (gob) encode all the proofs
	proofs := make([][]byte, len(blocks))
	proofLen := 0
	for i, b := range blocks {
		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)
		p, err := tree.GenerateProof(b)
		if err != nil {
			log.Fatalf("impossible to generate proof for block %v: %v", b, err)
		}
		if err = enc.Encode(p); err != nil {
			log.Fatal("encode:", err)
		}
		// the encoding of proofs should be fixed length and after some
		// tests it seems that gob encoding respect this
		proofs[i] = buff.Bytes()
		proofLen = len(proofs[i])
	}

	// enlarge the database, i.e., add the proof for every block
	enEntries := make([][]byte, len(db.Entries))
	p := 0
	bl := blockLenBytes
	for i := range db.Entries {
		enEntries[i] = make([]byte, 0)
		for j := 0; j < len(db.Entries[0])-bl; j += bl {
			enEntries[i] = append(enEntries[i], append(db.Entries[i][:j+bl], proofs[p]...)...)
			p++
		}
	}

	m := &Merkle{
		Entries:     enEntries,
		Info:        Info{},
		Root:        root,
		ProofLength: proofLen,
	}

	return m
}

func entriesToBlocks(e [][]byte, blockLength int) [][]byte {
	blocks := make([][]byte, len(e)*len(e[0])/blockLength)
	b := 0
	for i := range e {
		for j := 0; j < len(e[0])-blockLength; j += blockLength {
			blocks[b] = e[i][j : j+blockLength]
			b++
		}
	}

	return blocks
}
