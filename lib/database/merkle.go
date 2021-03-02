package database

import (
	"encoding/binary"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/merkle"
)

// CreateRandomMultiBitMerkle
// blockLen is the number of byte in a block, as byte is viewd as an element in this
// case
func CreateRandomMultiBitMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	entries := make([][]byte, numRows)
	numBlocks := dbLen / (8 * blockLen)
	// generate random blocks
	blocks := make([][]byte, numBlocks)
	for i := range blocks {
		// generate random block
		b := make([]byte, blockLen)
		if _, err := rnd.Read(b); err != nil {
			log.Fatal(err)
		}
		blocks[i] = b
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}

	// generate db
	blocksPerRow := numBlocks / numRows
	proofLen := 0
	b := 0
	for i := range entries {
		e := make([]byte, 0)
		for j := 0; j < blocksPerRow; j++ {
			p, err := tree.GenerateProof(blocks[b], 0)
			encodedProof := encodeProof(p)
			if err != nil {
				log.Fatalf("error while generating proof for block %v: %v", b, err)
			}
			e = append(e, append(blocks[b], encodedProof...)...)
			proofLen = len(encodedProof) // always same length
			b++
		}
		entries[i] = e
	}
	//db := CreateRandomMultiBitBytes(rnd, dbLen, numRows, blockLen)

	//// a leaf contains a block, which has the same numbers of bytes as a
	//// block composed of field elements
	//blocks := entriesToBlocks(db.Entries, blockLen)
	//tree, err := merkletree.New(blocks)
	//if err != nil {
	//log.Fatalf("impossible to create Merkle tree: %v", err)
	//}

	// get the root hash of the tree
	root := tree.Root()

	// generate and (gob) encode all the proofs
	//proofs := make([][]byte, len(blocks))
	//for i, b := range blocks {
	//p, err := tree.GenerateProof(b)
	//if err != nil {
	//log.Fatalf("error while generating proof for block %v: %v", b, err)
	//}
	//proofs[i] = encodeProof(p)
	//proofLen = len(proofs[i]) // always same length
	//}

	// enlarge the database, i.e., add the proof for every block
	//ee := make([][]byte, len(db.Entries))
	//p := 0
	//bl := blockLen
	//for i := range db.Entries {
	//ee[i] = make([]byte, 0)
	//for j := 0; j < len(db.Entries[0])-bl; j += bl {
	////ee[i] = append(ee[i], append(db.Entries[i][j:j+bl], proofs[p]...)...)
	//ee[i] = append(ee[i], append(blocks[p], proofs[p]...)...)
	//p++
	//}
	//}

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: dbLen / (8 * numRows * blockLen),
			BlockSize:  blockLen + proofLen,
			PIRType:    "merkle",
			Root:       root,
			ProofLen:   proofLen,
		},
	}

	return m
}

func DecodeProof(p []byte) *merkle.Proof {
	// number of hashes
	numHashes := binary.LittleEndian.Uint32(p[0:])

	// hashes
	hashLength := uint32(32) // sha256
	hashes := make([][]byte, numHashes)
	for i := uint32(0); i < numHashes; i++ {
		hashes[i] = p[4+hashLength*i : 4+hashLength*(i+1)]
	}

	// index
	index := binary.LittleEndian.Uint64(p[len(p)-8:])

	return &merkle.Proof{
		Hashes: hashes,
		Index:  index,
	}
}

func encodeProof(p *merkle.Proof) []byte {
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
