package database

import (
	"github.com/si-co/vpir-code/lib/merkle"
	"io"
	"log"
	"runtime"
)

// CreateRandomMerkle
// blockLen is the number of byte in a block,
// as byte is viewed as an element in this case
func CreateRandomMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
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

	// GC after tree generation
	runtime.GC()

	// generate db
	numColumns := numBlocks / numRows
	proofLen := tree.EncodedProofLength()
	// +1 is for storing the padding signal byte
	blockLen = blockLen + proofLen + 1
	blockLens := make([]int, numRows*numColumns)
	for b := 0; b < numRows*numColumns; b++ {
		blockLens[b] = blockLen
	}

	entries := makeMerkleEntries(blocks, tree, numRows, numColumns, blockLen)

	// GC after db creation
	runtime.GC()

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

func makeMerkleEntries(blocks [][]byte, tree *merkle.MerkleTree, nRows, nColumns, blockLen int) []byte {
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
		replyTo := make(chan []byte, 1)
		replies[i] = replyTo
		generateMerkleProofs(blocks[begin:end], tree, blockLen, replyTo)
	}

	for j, reply := range replies {
		chunk := <-reply
		output = append(output, chunk...)
		close(replies[j])
	}

	return output
}

func generateMerkleProofs(data [][]byte, t *merkle.MerkleTree, blockLen int, reply chan<- []byte) {
	result := make([]byte, 0, blockLen*len(data))
	//fmt.Println(len(data), blockLen, len(result))
	for b := 0; b < len(data); b++ {
		p, err := t.GenerateProof(data[b])
		if err != nil {
			log.Fatalf("error while generating proof for block %v: %v", b, err)
		}
		encodedProof := merkle.EncodeProof(p)
		// appending 0x80
		encodedProof = PadWithSignalByte(encodedProof)
		//encodedProof = PadWithSignalByte(encodedProof)
		// copying the data block and encoded proof into output
		result = append(result, append(data[b], encodedProof...)...)
		//copy(result[b*blockLen:(b+1)*blockLen], append(data[b], encodedProof...))
	}
	reply <- result
}
