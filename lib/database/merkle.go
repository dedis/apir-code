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
	for i, b := range blocks {
		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)
		p, err := tree.GenerateProof(b)
		if err != nil {
			log.Fatalf("error while generating proof for block %v: %v", b, err)
		}
		if err = enc.Encode(p); err != nil {
			log.Fatal("encode:", err)
		}
		// the encoding of proofs should be fixed length and after some
		// tests it seems that gob encoding respect this
		proofs[i] = buff.Bytes()
	}

	// pad gob encoding to fixed size
	max := 0
	for _, p := range proofs {
		if len(p) > max {
			max = len(p)
		}
	}
	proofLen := max
	fmt.Println("max:", proofLen)

	for i, p := range proofs {
		proofs[i] = PadBlock(p, proofLen)
		fmt.Println(len(proofs[i]))
	}

	// enlarge the database, i.e., add the proof for every block
	ee := make([][]byte, len(db.Entries))
	p := 0
	bl := blockLenBytes
	for i := range db.Entries {
		ee[i] = make([]byte, 0)
		for j := 0; j < len(db.Entries[0])-bl; j += bl {
			ee[i] = append(ee[i], append(db.Entries[i][:j+bl], proofs[p]...)...)
			p++
		}
		//fmt.Println(len(db.Entries[i]))
		//fmt.Println(len(ee[i]))
	}

	m := &Merkle{
		Entries: ee,
		Info: Info{
			NumRows:    numRows,
			NumColumns: 0,
			BlockSize:  blockLenBytes + proofLen,
		},
		Root:        root,
		ProofLength: proofLen,
	}

	return m
}

func entriesToBlocks(e [][]byte, blockLength int) [][]byte {
	blocks := make([][]byte, 0)
	var block []byte
	for i := range e {
		for j := 0; j < len(e[0]); j += blockLength {
			end := j + blockLength
			if end > len(e[0]) {
				// pad here with 0 bytes. We don't care about
				// the padding as this is not used in the PoC
				padLength := end - len(e[0])
				pad := make([]byte, padLength)
				block = append(e[i][j:end-padLength], pad...)
			} else {
				block = e[i][j:end]
			}
			blocks = append(blocks, block)
		}
	}

	return blocks
}
