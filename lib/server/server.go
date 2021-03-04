package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"math"
	"runtime"
	"sync"
)

// Server is a scheme-agnostic VPIR server interface, implemented by both IT
// and DPF-based schemes
type Server interface {
	AnswerBytes([]byte) ([]byte, error)
	DBInfo() *database.Info
}

// Answer computes the answer for the given query
func answer(q []field.Element, db *database.DB) []field.Element {
	// Doing simplified scheme if block consists of a single bit
	if db.BlockSize == cst.SingleBitBlockLength {
		a := make([]field.Element, db.NumRows)
		for i := 0; i < db.NumRows; i++ {
			for j := 0; j < db.NumColumns; j++ {
				if db.Entries[i*db.NumColumns+j].Equal(&cst.One) {
					a[i].Add(&a[i], &q[j])
				}
			}
		}
		return a
	}

	// parse the query
	qZeroBase := make([]field.Element, db.NumColumns)
	qOne := make([]field.Element, db.NumColumns*db.BlockSize)
	for j := 0; j < db.NumColumns; j++ {
		qZeroBase[j] = q[j*(db.BlockSize+1)]
		copy(qOne[j*db.BlockSize:(j+1)*db.BlockSize], q[j*(db.BlockSize+1)+1:(j+1)*(db.BlockSize+1)])
	}

	// multithreading
	numCores := runtime.NumCPU()
	//numCores := 2
	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	m := make([]field.Element, db.NumRows*(db.BlockSize+1))
	// we have to traverse column by column
	var begin, end int
	if db.NumRows == 1 {
		numWorkers := 0
		// channel to pass the ch from the routines back
		ch := make(chan []field.Element, numCores*(db.BlockSize+1))
		colPerChunk := divideAndCeil(db.NumColumns, numCores)
		for j := 0; j < db.NumColumns; j += colPerChunk {
			begin, end = computeChunkIndices(j, colPerChunk, db.NumColumns, db.BlockSize)
			go processRowChunk(db.Entries[begin:end], db.BlockSize, qZeroBase[j:j+colPerChunk], qOne[begin:end], ch)
			numWorkers++
		}
		result := combineChunkResults(numWorkers, db.BlockSize+1, ch)
		copy(m, result)
		close(ch)
	} else {
		var wg sync.WaitGroup
		rowsPerCore := divideAndCeil(db.NumRows, numCores)
		for j := 0; j < db.NumRows; j += rowsPerCore {
			begin, end = computeChunkIndices(j, rowsPerCore, db.NumRows, db.BlockSize)
			wg.Add(1)
			go processRows(db.Entries[begin*db.NumColumns: end*db.NumColumns], db.BlockSize, qZeroBase, qOne, &wg, m[begin:end])
		}
		wg.Wait()
	}

	return m
}

func processRows(rows []field.Element, blockLen int, qZ []field.Element, qO []field.Element, wg *sync.WaitGroup, output []field.Element) {
	numElementsInRow := len(qO)
	for i := 0; i < len(rows)/numElementsInRow; i++ {
		res := multiplyAndTag(rows[i*numElementsInRow:(i+1)*numElementsInRow], blockLen, qZ, qO)
		copy(output[i*blockLen:(i+1)*blockLen], res)
	}
	wg.Done()
}

// processing a chunk of a database row
func processRowChunk(chunk []field.Element, blockLen int, qZ []field.Element, qO []field.Element, reply chan<- []field.Element) {
	reply <- multiplyAndTag(chunk, blockLen, qZ, qO)
}

// combine the results of processing a row by different routines
func combineChunkResults(nw int, resLen int, workerReplies <-chan []field.Element) []field.Element {
	product := make([]field.Element, resLen)
	for i := 0; i < nw; i++ {
		reply := <-workerReplies
		for i, elem := range reply {
			product[i].Add(&product[i], &elem)
		}
	}
	return product
}

// multiplyAndTag multiplies db entries with the elements
// from the client query and computes a tag over each block
func multiplyAndTag(elements []field.Element, blockLen int, tagBase []field.Element, messageBase []field.Element) []field.Element {
	var prodTag, prod field.Element
	sumTag := field.Zero()
	sum := field.ZeroVector(blockLen)
	for j := 0; j < len(elements)/blockLen; j++ {
		for b := 0; b < blockLen; b++ {
			if elements[j*blockLen+b].IsZero() {
				// no need to multiply if the element value is zero
				continue
			}
			// compute message
			prod.Mul(&elements[j*blockLen+b], &tagBase[j])
			sum[b].Add(&sum[b], &prod)
			// compute block tag
			prodTag.Mul(&elements[j*blockLen+b], &messageBase[j*blockLen+b])
			sumTag.Add(&sumTag, &prodTag)
		}
	}
	return append(sum, sumTag)
}

func computeChunkIndices(ind, step, max, multiplier int) (int, int) {
	// avoiding overflow when colPerChunk does not divide db.Columns evenly
	if ind+step > max {
		step = max - ind
	}
	return ind*multiplier, (ind+step)*multiplier
}

func divideAndCeil(dividend, divisor int) int {
	return int(math.Ceil(float64(dividend) / float64(divisor)))
}
