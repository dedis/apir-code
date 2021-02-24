package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/field"
	merkletree "github.com/wealdtech/go-merkletree"
)

type Merkle struct {
	DB          *Bytes
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

	// getch the root hash of the tree
	//root := tree.Root()

	// generate and (gob) encode all the proofs
	proofs := make([][]byte, len(blocks))
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
		// tests it seems that gob encoding respect that
		proofs[i] = buff.Bytes()
		fmt.Println(len(proofs[i]))

	}

	// enlarge the database, i.e., add the proof for every block

	//// hash function for the Merkle tree
	//h := sha256.New()

	//// compute the proof for each block
	//var mr []byte
	//proofs := make([][]byte, len(entriesFlatten))
	//for i, _ := range entriesFlatten {
	//r := bytes.NewReader(entriesFlatten)
	//merkleRoot, proof, _, err := merkletree.BuildReaderProof(r, h, segmentSize, uint64(i))
	//if err != nil {
	//panic(err)
	//}
	//fmt.Println(proof)
	//proofs[i] = flatten(proof)
	//mr = merkleRoot // always the same
	//}

	//m := &Merkle{
	//DB:         db,
	//MerkleRoot: mr,
	//}

	//// enlarge the db by adding a proof to each block
	//enlargedEntries := make([][]byte, 0) // TODO: specify length
	//for i := range db.Entries {
	//for j := 0; j < db.Entries[0]; j += segmentSize {
	//a = append(a[:i], append(make([]T, j), a[i:]...)...)
	//append(enlargedEntries[i][:j], append(proofs[i*len(db.Entries[0])+j], enlar
	//}
	//}

	//return m
	return nil
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
