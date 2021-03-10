package database

import (
	"io"
	"log"
	"runtime"
	"sync"

	"github.com/si-co/vpir-code/lib/merkle"
	"github.com/si-co/vpir-code/lib/utils"
)

// CreateRandomMultiBitMerkle
// blockLen is the number of byte in a block,
// as byte is viewed as an element in this case
func CreateRandomMultiBitMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	numBlocks := dbLen / (8 * blockLen)
	// generate random blocks blocks
	data := make([]byte, numBlocks*blockLen)
	if _, err := rnd.Read(data); err != nil {
		log.Fatal(err)
	}

	blocks := make([][]byte, numBlocks)
	for i := range blocks {
		// generate random block
		blocks[i] = make([]byte, blockLen)
		copy(blocks[i], data[i*blockLen:(i+1)*blockLen])
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}

	// generate db
	numColumns := numBlocks / numRows
	proofLen := tree.EncodedProofLength()
	blockLen = blockLen + proofLen
	entries := make([]byte, numRows * numColumns * blockLen)
	// multithreading
	var wg sync.WaitGroup
	var end int
	numCores := runtime.NumCPU()
	chunkLen := utils.DivideAndRoundUp(numRows*numColumns, numCores)
	for i := 0; i < numRows*numColumns; i += chunkLen {
		end = i + chunkLen
		if end > numRows*numColumns {
			end = numRows*numColumns
		}
		wg.Add(1)
		go assignEntries(entries[i*blockLen:end*blockLen], blocks[i:end], tree, blockLen, &wg)
	}
	wg.Wait()

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: numColumns,
			BlockSize:  blockLen,
			PIRType:    "merkle",
			Root:       tree.Root(),
			ProofLen:   proofLen,
		},
	}

	return m
}

func assignEntries(es []byte, bks [][]byte, t *merkle.MerkleTree, blockLen int, wg *sync.WaitGroup) {
	for b := 0; b < len(bks); b++ {
		p, err := t.GenerateProof(bks[b])
		if err != nil {
			log.Fatalf("error while generating proof for block %v: %v", b, err)
		}
		encodedProof := merkle.EncodeProof(p)
		copy(es[b*blockLen:(b+1)*blockLen], append(bks[b], encodedProof...))
	}
	wg.Done()
}
