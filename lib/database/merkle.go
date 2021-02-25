package database

import (
	"encoding/binary"
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
		p, err := tree.GenerateProof(b)
		if err != nil {
			log.Fatalf("error while generating proof for block %v: %v", b, err)
		}
		proofs[i] = encodeProof(p)
		proofLen = len(proofs[i]) // always the same
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
	}

	m := &Merkle{
		Entries: ee,
		Info: Info{
			NumRows:    numRows,
			NumColumns: len(ee[0]),
			BlockSize:  blockLenBytes + proofLen,
		},
		Root:        root,
		ProofLength: proofLen,
	}

	return m
}

func encodeProof(p *merkletree.Proof) []byte {
	out := make([]byte, 0)

	// encode number of hashes
	numHashes := uint32(len(p.Hashes))
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, numHashes)
	out = append(out, b...)

	// encode hashes
	for _, h := range p.Hashes {
		out = append(out, h...)
	}

	// encode index
	b1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b1, p.Index)
	out = append(out, b1...)

	return out
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
