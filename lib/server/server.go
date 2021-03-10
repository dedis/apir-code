package server

import (
	"math"
	"runtime"
	"sync"

	"github.com/lukechampine/fastxor"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
)

// Server is a scheme-agnostic VPIR server interface, implemented by both IT
// and DPF-based schemes
type Server interface {
	AnswerBytes([]byte) ([]byte, error)
	DBInfo() *database.Info
}

/*
%%	VPIR primitives
*/

// Answer computes the VPIR answer for the given query
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

	// multithreading
	numCores := runtime.NumCPU()
	//numCores := 2
	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	// we have to traverse column by column
	var begin, end int
	if db.NumRows == 1 {
		numWorkers := 0
		// channel to pass the ch from the routines back
		ch := make(chan []field.Element, numCores*(db.BlockSize+1))
		colPerChunk := divideAndRoundUp(db.NumColumns, numCores)
		for j := 0; j < db.NumColumns; j += colPerChunk {
			colPerChunk, begin, end = computeChunkIndices(j, colPerChunk, db.NumColumns, db.BlockSize)
			go processRowChunk(db.Entries[begin:end], db.BlockSize, q[j*(db.BlockSize+1):(j+colPerChunk)*(db.BlockSize+1)], ch)
			numWorkers++
		}
		result := combineChunkResults(numWorkers, db.BlockSize+1, ch)
		close(ch)
		return result
	} else {
		m := make([]field.Element, db.NumRows*(db.BlockSize+1))
		var wg sync.WaitGroup
		rowsPerCore := divideAndRoundUp(db.NumRows, numCores)
		for j := 0; j < db.NumRows; j += rowsPerCore {
			rowsPerCore, begin, end = computeChunkIndices(j, rowsPerCore, db.NumRows, db.BlockSize)
			wg.Add(1)
			go processRows(db.Entries[begin*db.NumColumns:end*db.NumColumns], db.BlockSize, db.NumColumns, q, &wg,
				m[j*(db.BlockSize+1):(j+rowsPerCore)*(db.BlockSize+1)])
		}
		wg.Wait()

		return m
	}
}

// processing multiple rows by iterating over them
func processRows(rows []field.Element, blockLen, numColumns int, q []field.Element, wg *sync.WaitGroup, output []field.Element) {
	numElementsInRow := blockLen * numColumns
	for i := 0; i < len(rows)/numElementsInRow; i++ {
		res := multiplyAndTag(rows[i*numElementsInRow:(i+1)*numElementsInRow], blockLen, q)
		copy(output[i*(blockLen+1):(i+1)*(blockLen+1)], res)
	}
	wg.Done()
}

// processing a chunk of a database row
func processRowChunk(chunk []field.Element, blockLen int, q []field.Element, reply chan<- []field.Element) {
	reply <- multiplyAndTag(chunk, blockLen, q)
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
func multiplyAndTag(elements []field.Element, blockLen int, q []field.Element) []field.Element {
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
			prod.Mul(&elements[j*blockLen+b], &q[j*(blockLen+1)])
			sum[b].Add(&sum[b], &prod)
			// compute block tag
			prodTag.Mul(&elements[j*blockLen+b], &q[j*(blockLen+1)+1+b])
			sumTag.Add(&sumTag, &prodTag)
		}
	}
	return append(sum, sumTag)
}

func computeChunkIndices(ind, step, max, multiplier int) (int, int, int) {
	// avoiding overflow when colPerChunk does not divide db.Columns evenly
	if ind+step > max {
		step = max - ind
	}
	return step, ind * multiplier, (ind + step) * multiplier
}

func divideAndRoundUp(dividend, divisor int) int {
	return int(math.Ceil(float64(dividend) / float64(divisor)))
}

/*
%%	PIR primitives
*/
func answerPIR(q []byte, db *database.Bytes) []byte {
	bs := db.BlockSize
	m := make([]byte, db.NumRows*bs)
	rs := db.NumColumns * bs
	// we have to traverse column by column
	for i := 0; i < db.NumRows; i++ {
		sum := make([]byte, bs)
		for j := 0; j < db.NumColumns; j++ {
			if q[j] == byte(1) {
				fastxor.Bytes(sum, sum, db.Entries[i*rs+j*bs:i*rs+(j+1)*bs])
			}
		}
		copy(m[i*bs:(i+1)*bs], sum)
	}

	return m
}
