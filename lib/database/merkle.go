package database

import (
	"fmt"
	"github.com/si-co/vpir-code/lib/merkle"
	"io"
	"log"
	"runtime"
)

// CreateRandomMultiBitMerkle
// blockLen is the number of byte in a block,
// as byte is viewed as an element in this case
func CreateRandomMultiBitMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	numBlocks := dbLen / (8 * blockLen)
	// generate random numBlocks blocks
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
	blockLens := make([]int, numRows*numColumns)
	for b := 0; b < numRows*numColumns; b++ {
		blockLens[b] = blockLen
	}

	entries := makeMerkleEntries(blocks, tree, numRows, numColumns, blockLen)
	fmt.Println(entries)

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:      numRows,
			NumColumns:   numColumns,
			BlockSize:    blockLen,
			BlockLengths: blockLens,
			PIRType:      "merkle",
			Merkle:       &Merkle{Root: tree.Root(), ProofLen: proofLen},
		},
	}

	return m
}

func makeMerkleEntries(blocks [][]byte, tree *merkle.MerkleTree, nRows, nColumns, maxBlockLen int) []byte {
	output := make([]byte, 0)
	var begin, end int
	NGoRoutines := runtime.NumCPU()
	replies := make([]chan []byte, NGoRoutines)
	blocksPerRoutine := nRows * nColumns / NGoRoutines
	for i := 0; i < NGoRoutines; i++ {
		begin, end = i*blocksPerRoutine, (i+1)*blocksPerRoutine
		if end > nRows*nColumns || i == NGoRoutines-1 {
			end = nRows * nColumns
		}
		replyTo := make(chan []byte, maxBlockLen*(end-begin))
		replies[i] = replyTo
		generateMerkleProofs(blocks[begin:end], tree, replyTo)
	}

	for j, reply := range replies {
		chunk := <-reply
		output = append(output, chunk...)
		close(replies[j])
	}

	return output
}

func generateMerkleProofs(data [][]byte, t *merkle.MerkleTree, reply chan<- []byte) {
	result := make([]byte, 0)
	for b := 0; b < len(data); b++ {
		p, err := t.GenerateProof(data[b])
		if err != nil {
			log.Fatalf("error while generating proof for block %v: %v", b, err)
		}
		encodedProof := merkle.EncodeProof(p)
		result = append(result, append(data[b], encodedProof...)...)
	}
	reply <- result
}
